using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.Networking;

using Postman;
using Newtonsoft.Json;
using Cysharp.Threading.Tasks;

public class PostmanHttpUtil
{
#region publish
    public static async UniTask<ResultMessageData> PublishAsync(string host, string channel, string message, string tag = "", string extention = "", bool useSSL = false)
    {
        string url = string.Format("{0}://{1}/postman/publish?ch={2}&msg={3}&tag={4}&ext={5}",
            (useSSL ? "https" : "http"),
            host,
            Uri.EscapeDataString(channel),
            Uri.EscapeDataString(message),
            Uri.EscapeDataString(tag),
            Uri.EscapeDataString(extention));

        UnityWebRequest request = UnityWebRequest.Get(url);
        await request.SendWebRequest();

        ResultMessageData responce;
        if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
        {
            Debug.LogError("PostmanHttpLib :: " + request.error);
            responce = new ResultMessageData("", request.error);
        }
        else
        {
            responce = JsonConvert.DeserializeObject<ResultMessageData>(request.downloadHandler.text);
            Debug.Log(string.Format("PostmanHttpLib :: publish [ {0} ] {1}", channel, message)
                + ((tag != "") ? " / " + tag : "")
                + ((extention != "") ? " / " + extention : ""));
        }

        request.Dispose();

        return responce;
    }

    public static async UniTask PublishWithRetryAsync(string host, string channel, string message, string tag = "", string extention = "", bool useSSL = false)
    {
        int retry = 5;

        ResultMessageData res = await PublishAsync(host, channel, message, tag, extention, useSSL);
        while(res.IsError() && retry > 0)
        {
            Debug.LogError(res.error);

            retry--;
            res = await PublishAsync(host, channel, message, tag, extention, useSSL);
        }
    }
#endregion

#region store
    public static async UniTask<ResultMessageData> StoreSetAsDataAsync(string host, string key, string val, bool useSSL = false)
    {
        string url = string.Format("{0}://{1}/postman/store?cmd=SET&key={2}&val={3}",
            (useSSL ? "https" : "http"),
            host,
            Uri.EscapeDataString(key),
            Uri.EscapeDataString(val));

        UnityWebRequest request = UnityWebRequest.Get(url);
        await request.SendWebRequest();

        ResultMessageData responce = new ResultMessageData("", "");
        if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
            Debug.LogError("PostmanHttpLib :: " + request.error);
        else
        {
            responce = JsonConvert.DeserializeObject<ResultMessageData>(request.downloadHandler.text);
            Debug.Log(string.Format("PostmanHttpLib :: store set [ {0} : {1} ]", key, val));
        }

        request.Dispose();

        return responce;
    }

    public static async UniTask StoreSetWithRetryAsync(string host, string key, string val, bool useSSL = false)
    {
        int retry = 5;

        ResultMessageData res = await StoreSetAsDataAsync(host, key, val, useSSL);
        while((res.result == "" || res.result != "success" || res.IsError())
            && retry > 0)
        {
            Debug.LogError(res.error);

            retry--;
            res = await StoreSetAsDataAsync(host, key, val, useSSL);
        }
    }

    public static async UniTask<ResultMessageData> StoreGetAsDataAsync(string host, string key, bool useSSL = false)
    {
        string url = string.Format("{0}://{1}/postman/store?cmd=GET&key={2}",
            (useSSL ? "https" : "http"),
            host,
            Uri.EscapeDataString(key));

        UnityWebRequest request = UnityWebRequest.Get(url);
        await request.SendWebRequest();

        ResultMessageData responce;
        if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
        {
            Debug.LogError("PostmanHttpLib :: " + request.error);
            responce = new ResultMessageData("", request.error);
        }
        else
        {
            responce = JsonConvert.DeserializeObject<ResultMessageData>(request.downloadHandler.text);
            Debug.Log(string.Format("PostmanHttpLib :: store get [ {0} : {1} ]", key, responce.result));
        }

        request.Dispose();

        return responce;
    }

    public static async UniTask<string> StoreGetWithRetryAsync(string host, string key, bool useSSL = false)
    {
        int retry = 5;

        ResultMessageData res = await StoreGetAsDataAsync(host, key, useSSL);
        while(res.IsError() && retry > 0)
        {
            Debug.LogError(res.error);

            retry--;
            res = await StoreGetAsDataAsync(host, key, useSSL);
        }

        return res.result;
    }
#endregion

#region status
    public static async UniTask<StatusMessageData> StatusAsync(string host, bool useSSL = false)
    {
        string url = string.Format("{0}://{1}/postman/status", (useSSL ? "https" : "http"), host);

        UnityWebRequest request = UnityWebRequest.Get(url);
        await request.SendWebRequest();

        StatusMessageData responce;
        if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
        {
            Debug.LogError("PostmanHttpLib :: " + request.error);
            responce = new StatusMessageData("", null, request.error);
        }
        else
        {
            responce = JsonConvert.DeserializeObject<StatusMessageData>(request.downloadHandler.text);
            Debug.Log(string.Format("PostmanHttpLib :: status: {0}", request.downloadHandler.text));
        }

        request.Dispose();

        return responce;
    }
#endregion
}
