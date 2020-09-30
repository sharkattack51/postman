using System.Collections;
using System.Collections.Generic;
using System;
using UnityEngine.Networking;
using UnityEngine;
using UnityEditor;
using WebSocketSharp;
using Newtonsoft.Json;

namespace Postman
{
    public class PostmanClient : MonoBehaviour
    {
        public string serverIp = "127.0.0.1:8800";
        public bool connectOnStart = true;
        public bool reconnectOnClose = false;
        [TextArea] public string secureToken = "";
        public Action OnConnect;
        public Action<PublishMessageData> OnMessage;
        public Action OnClose;
        public Action OnPingPong;

        private WebSocket webSocket;

        private bool isConnect = false;
        public bool IsConnect { get{ return isConnect; } }

        private PublishMessageData latestMessage = null;
        public PublishMessageData LatestMessage { get{ return latestMessage; } }

        private List<PublishMessageData> messageStack = new List<PublishMessageData>();

        private bool tryReconnect = false;
        private bool reconnecting = false;

        private bool invokeOnConnect = false;
        private bool invokeOnClose = false;
        private bool invokePing = false;


        void Awake()
        {

        }

        void Start()
        {
#if UNITY_EDITOR && UNITY_2018_3_OR_NEWER
            if(PlayerSettings.scriptingRuntimeVersion != ScriptingRuntimeVersion.Latest)
                Debug.LogError("PostmanClient :: PlayerSettings.scriptingRuntimeVersion is Lagacy");

            BuildTargetGroup target = BuildPipeline.GetBuildTargetGroup(EditorUserBuildSettings.activeBuildTarget);
            if(PlayerSettings.GetApiCompatibilityLevel(target) != ApiCompatibilityLevel.NET_4_6)
                Debug.LogError("PostmanClient :: PlayerSettings.ApiCompatibilityLevel is Low");
#endif

            if(connectOnStart)
                Connect(true);
        }

        void Update()
        {
            if(tryReconnect)
            {
                reconnecting = true;
                StartCoroutine(ReconnectCoroutine());
                tryReconnect = false;
            }

            if(invokeOnConnect)
            {
                if(OnConnect != null)
                    OnConnect();

                invokeOnConnect = false;
            }

            if(messageStack != null && messageStack.Count > 0)
            {
                PublishMessageData[] copyStack;

                lock(((ICollection)messageStack).SyncRoot)
                {
                    copyStack = messageStack.ToArray();
                    messageStack = new List<PublishMessageData>();
                }

                foreach(PublishMessageData msg in copyStack)
                {
                    Debug.Log(string.Format("PostmanClient :: [{0}] > {1}", msg.channel, msg.message));

                    if(OnMessage != null)
                        OnMessage(msg);

                    latestMessage = msg;
                }
            }

            if(invokeOnClose)
            {
                if(OnClose != null)
                    OnClose();

                invokeOnClose = false;
            }

            if(invokePing)
            {
                Debug.Log("PostmanClient :: pong");

                if(OnPingPong != null)
                    OnPingPong();

                invokePing = false;
            }
        }

        void OnDestroy()
        {
            reconnectOnClose = false;
            Close(true);
        }


#region connect/close
        public void Connect(bool asAsync = false)
        {
            if(isConnect && webSocket != null && webSocket.IsAlive)
            {
                Debug.Log("PostmanClient :: already connected");
                return;
            }

            if(webSocket != null)
                Close(true);

            try
            {
                string ip = serverIp.Replace("http://", "").Replace("https://", "");
                string url = string.Format("ws://{0}/postman", ip);
                if(secureToken != "")
                    url += "?tkn=" + secureToken;

                webSocket = new WebSocket(url);
                webSocket.Compression = CompressionMethod.Deflate;
                webSocket.OnOpen += OnWebSocketOpen;
                webSocket.OnMessage += OnWebSocketMessage;
                webSocket.OnClose += OnWebSocketClose;
                webSocket.OnError += OnWebSocketError;
                if(asAsync)
                    webSocket.ConnectAsync();
                else
                    webSocket.Connect();
            }
            catch(Exception e)
            {
                Debug.LogError("PostmanClient :: connection error - " + e.Message);
            }
        }

        public void Close(bool asAsync = false)
        {
            isConnect = false;

            if(webSocket != null)
            {
                webSocket.OnOpen -= OnWebSocketOpen;
                webSocket.OnMessage -= OnWebSocketMessage;
                webSocket.OnClose -= OnWebSocketClose;
                webSocket.OnError -= OnWebSocketError;
                if(asAsync)
                    webSocket.CloseAsync();
                else
                    webSocket.Close();
            }
            webSocket = null;
        }

        private IEnumerator ReconnectCoroutine()
        {
            while(true)
            {
                Connect(true);

                yield return new WaitForSeconds(1.0f);

                if(isConnect && webSocket != null && webSocket.IsAlive)
                    break;
            }

            reconnecting = false;
        }
#endregion

#region ws event function
        private void OnWebSocketOpen(object sender, EventArgs e)
        {
            Debug.Log("PostmanClient :: client opened");

            isConnect = true;
            invokeOnConnect = true;
        }

        private void OnWebSocketMessage(object sender, MessageEventArgs e)
        {
            if(!e.IsText || e.Data == "")
                return;

            if(!e.Data.Contains(PostmanMassageData.ProtocolMessageTag))
                return;

            string msgstr = e.Data.Split(new string[]{ PostmanMassageData.ProtocolMessageTag }, StringSplitOptions.None)[1];
            if(msgstr == PostmanMassageData.PingReturnString)
                invokePing = true;
            else
            {
                try
                {
                    PublishMessageData msg = JsonConvert.DeserializeObject<PublishMessageData>(msgstr);

                    lock(((ICollection)messageStack).SyncRoot)
                        messageStack.Add(msg);
                }
                catch
                {
                    return;
                }
            }			
        }

        private void OnWebSocketClose(object sender, CloseEventArgs e)
        {
            Debug.Log("PostmanClient :: client closed");

            isConnect = false;
            invokeOnClose = true;

            if(reconnectOnClose && !reconnecting)
                tryReconnect = true;
        }

        private void OnWebSocketError(object sender, ErrorEventArgs e)
        {
            Debug.LogError("PostmanClient :: client error - " + e.Exception);
        }
#endregion

#region postman send function
        public void Ping()
        {
            if(isConnect && webSocket != null)
                StartCoroutine(PingCoroutine());
        }

#region ping
        private IEnumerator PingCoroutine()
        {
            webSocket.SendAsync(PostmanMassageData.BuildMessage(MessageType.PING), null);
            yield return null;
        }
#endregion

#region subscribe
        public void Subscribe(string channel, string client_info = "")
        {
            if(isConnect && webSocket != null && webSocket.IsAlive)
                StartCoroutine(SubscribeCoroutine(channel, client_info));
        }

        private IEnumerator SubscribeCoroutine(string channel, string client_info)
        {
            SubscribeMessageData sub = new SubscribeMessageData(channel, client_info);
            string json = JsonConvert.SerializeObject(sub);
            webSocket.SendAsync(PostmanMassageData.BuildMessage(MessageType.SUBSCRIBE, json), null);
            yield return null;
        }
#endregion

#region unsubscribe
        public void Unsubscribe(string channel)
        {
            if(isConnect && webSocket != null && webSocket.IsAlive)
                StartCoroutine(UnsubscribeCoroutine(channel));
        }

        private IEnumerator UnsubscribeCoroutine(string channel)
        {
            SubscribeMessageData sub = new SubscribeMessageData(channel);
            string json = JsonConvert.SerializeObject(sub);
            webSocket.SendAsync(PostmanMassageData.BuildMessage(MessageType.UNSUBSCRIBE, json), null);
            yield return null;
        }
#endregion

#region publish
        public void Publish(string channel, string message, string tag = "", string extention = "")
        {
            if(isConnect && webSocket != null && webSocket.IsAlive)
                StartCoroutine(PublishCoroutine(channel, message, tag, extention));
        }

        private IEnumerator PublishCoroutine(string channel, string message, string tag, string extention)
        {
            PublishMessageData pub = new PublishMessageData(channel, message, tag, extention);
            string json = JsonConvert.SerializeObject(pub);

            if(Application.platform == RuntimePlatform.Android)
                webSocket.Send(PostmanMassageData.BuildMessage(MessageType.PUBLISH, json));
            else
                webSocket.SendAsync(PostmanMassageData.BuildMessage(MessageType.PUBLISH, json), null);

            yield return null;
        }
#endregion

#region store get
        public string StoreGet(string key)
        {
            ResultMessageData data = StoreGetAsData(key);
            return data.result;
        }

        public ResultMessageData StoreGetAsData(string key)
        {
            string ip = serverIp.Replace("http://", "").Replace("https://", "");
            string url = string.Format("http://{0}/postman/store?cmd=GET&key={1}", ip, key);
            if(secureToken != "")
                url += "&tkn=" + secureToken;

            UnityWebRequest request = UnityWebRequest.Get(url);
            request.SendWebRequest();

            while(!request.isDone);

            ResultMessageData responce;
            if(request.isHttpError || request.isNetworkError)
            {
                Debug.LogError("PostmanClient :: " + request.error);
                responce = new ResultMessageData("", request.error);
            }
            else
            {
                responce = JsonConvert.DeserializeObject<ResultMessageData>(request.downloadHandler.text);
                Debug.Log(string.Format("PostmanClient :: store get [ {0} : {1} ]", key, responce.result));
            }

            request.Dispose();

            return responce;
        }
#endregion

#region store set
        public void StoreSet(string key, string val)
        {
            string ip = serverIp.Replace("http://", "").Replace("https://", "");
            string url = string.Format("http://{0}/postman/store?cmd=SET&key={1}&val={2}", ip, key, val);
            if(secureToken != "")
                url += "&tkn=" + secureToken;

            UnityWebRequest request = UnityWebRequest.Get(url);
            request.SendWebRequest();

            while(!request.isDone);

            if(request.isHttpError || request.isNetworkError)
                Debug.LogError("PostmanClient :: " + request.error);
            else
                Debug.Log(string.Format("PostmanClient :: store set [ {0} : {1} ]", key, val));

            request.Dispose();
        }
#endregion

#region store haskey
        public bool StoreHasKey(string key)
        {
            ResultMessageData data = StoreHasKeyAsData(key);

            bool b = false;
            return bool.TryParse(data.result, out b);
        }

        public ResultMessageData StoreHasKeyAsData(string key)
        {
            string ip = serverIp.Replace("http://", "").Replace("https://", "");
            string url = string.Format("http://{0}/postman/store?cmd=HAS&key={1}", ip, key);
            if(secureToken != "")
                url += "&tkn=" + secureToken;

            UnityWebRequest request = UnityWebRequest.Get(url);
            request.SendWebRequest();

            while(!request.isDone);

            ResultMessageData responce;
            if(request.isHttpError || request.isNetworkError)
            {
                Debug.LogError("PostmanClient :: " + request.error);
                responce = new ResultMessageData("", request.error);
            }
            else
            {
                responce = JsonConvert.DeserializeObject<ResultMessageData>(request.downloadHandler.text);
                Debug.Log(string.Format("PostmanClient :: store haskey [ {0} ] ", key) + responce.result);
            }

            request.Dispose();

            return responce;
        }
#endregion

#region store delete
        public void StoreDelete(string key)
        {
            string ip = serverIp.Replace("http://", "").Replace("https://", "");
            string url = string.Format("http://{0}/postman/store?cmd=DEL&key={1}", ip, key);
            if(secureToken != "")
                url += "&tkn=" + secureToken;

            UnityWebRequest request = UnityWebRequest.Get(url);
            request.SendWebRequest();

            while(!request.isDone);

            if(request.isHttpError || request.isNetworkError)
                Debug.LogError("PostmanClient :: " + request.error);
            else
                Debug.Log(string.Format("PostmanClient :: store delete [ {0} ]", key));

            request.Dispose();
        }
#endregion

#endregion
    }
}
