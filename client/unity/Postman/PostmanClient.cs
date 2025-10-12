using System;
using System.Collections;
using System.Collections.Generic;
using System.Net;
using System.Net.NetworkInformation;
using System.Threading;
using UnityEngine;
using UnityEngine.Networking;

using WebSocketSharp;
using Newtonsoft.Json;
using Cysharp.Threading.Tasks;

using Ping = System.Net.NetworkInformation.Ping;

namespace Postman
{
    public class PostmanClient : MonoBehaviour
    {
        [Header("connect setting")]
        public string serverIpOrUrl = "127.0.0.1:8800";
        public bool useSSL = false;
        public bool connectOnStart = true;

        private string host = "";

        [Header("connect retry setting")]
        public bool reconnectOnClose = true;
        public bool exponentialBackoff = false;
        public bool checkNetworkReachable = false;

        [Header("for secure mode option")]
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

        private bool isReachablePing = true;
        private CancellationTokenSource reachablePingCts;

        public static int CHECK_NETWORK_REACHABLE_SPAN_MSEC = 3000;


        void Awake()
        {

        }

        void Start()
        {
#if UNITY_EDITOR && UNITY_2018_3_OR_NEWER

#if !UNITY_2019_3_OR_NEWER
            if(PlayerSettings.scriptingRuntimeVersion != ScriptingRuntimeVersion.Latest)
                Debug.LogError("PostmanClient :: PlayerSettings.scriptingRuntimeVersion is Lagacy");
#endif

#if !UNITY_2020_1_OR_NEWER
            BuildTargetGroup target = BuildPipeline.GetBuildTargetGroup(EditorUserBuildSettings.activeBuildTarget);
            if(PlayerSettings.GetApiCompatibilityLevel(target) != ApiCompatibilityLevel.NET_4_6)
                Debug.LogError("PostmanClient :: PlayerSettings.ApiCompatibilityLevel is Low");
#endif

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

            // WebSocket.OnClose was not called when the wired LAN was disconnected.
            if(checkNetworkReachable)
            {
                if(isConnect && !isReachablePing)
                {
                    Close(true);
                    OnWebSocketClose(null, null);
                }
            }
        }

        void OnDestroy()
        {
            reconnectOnClose = false;
            Close(true);
        }


#region connect/close
        public void Connect(bool asAsync = true)
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
                serverIpOrUrl = serverIpOrUrl.Trim();
                serverIpOrUrl = serverIpOrUrl.TrimEnd('/');

                host = "";
                if(serverIpOrUrl.Contains("http://"))
                {
                    useSSL = false;
                    host = serverIpOrUrl.Replace("http://", "");
                }
                else if(serverIpOrUrl.Contains("https://"))
                {
                    useSSL = true;
                    host = serverIpOrUrl.Replace("https://", "");
                }
                else if(serverIpOrUrl.Contains("ws://"))
                {
                    useSSL = false;
                    host = serverIpOrUrl.Replace("ws://", "");
                }
                else if(serverIpOrUrl.Contains("wss://"))
                {
                    useSSL = true;
                    host = serverIpOrUrl.Replace("wss://", "");
                }
                else
                    host = serverIpOrUrl;

                host = host.Replace("/postman", "");
                string url = host + "/postman";
                if(useSSL)
                    url = "wss://" + url;
                else
                    url = "ws://" + url;

                if(secureToken != "")
                    url += "?tkn=" + secureToken;

                webSocket = new WebSocket(url);
                if(useSSL)
                    webSocket.SslConfiguration.EnabledSslProtocols = System.Security.Authentication.SslProtocols.Tls12;
                webSocket.Compression = CompressionMethod.Deflate;
                webSocket.OnOpen += OnWebSocketOpen;
                webSocket.OnMessage += OnWebSocketMessage;
                webSocket.OnClose += OnWebSocketClose;
                webSocket.OnError += OnWebSocketError;
                if(asAsync)
                    webSocket.ConnectAsync();
                else
                    webSocket.Connect();

                // WebSocket.OnClose was not called when the wired LAN was disconnected.
                if(checkNetworkReachable)
                {
                    isReachablePing = true;
                    reachablePingCts = new CancellationTokenSource();
                    UniTask.RunOnThreadPool(async (tkn) =>
                    {
                        await UniTask.WaitUntil(() => webSocket != null && webSocket.IsAlive);

                        while(reachablePingCts != null && !reachablePingCts.IsCancellationRequested && isReachablePing)
                        {
                            await UniTask.Delay(CHECK_NETWORK_REACHABLE_SPAN_MSEC);

                            try
                            {
                                Ping ping = new Ping();
                                PingReply reply = null;

                                string ip = serverIpOrUrl.Split(new string[] { ":" }, StringSplitOptions.None)[0];
                                IPAddress ipa;
                                if(IPAddress.TryParse(ip, out ipa))
                                    reply = await ping.SendPingAsync(ipa, 2000);
                                else
                                {
                                    string host = serverIpOrUrl.Replace("http://", "").Replace("https://", "").Replace("ws://", "").Replace("wss://", "");
                                    IPHostEntry hostEntry = Dns.GetHostEntry(host);
                                    IPAddress addr = null;
                                    foreach(IPAddress a in hostEntry.AddressList)
                                    {
                                        if(a.AddressFamily == System.Net.Sockets.AddressFamily.InterNetwork) // IPv4
                                        {
                                            addr = a;
                                            break;
                                        }
                                    }
                                    if(addr != null)
                                    {
                                        PingOptions opt = new PingOptions(128, true);
                                        byte[] buf = new byte[32];
                                        reply = await ping.SendPingAsync(addr, 2000, buf, opt);
                                    }
                                }

                                if(reply != null && reply.Status != IPStatus.Success)
                                    isReachablePing = false;
                            }
                            catch
                            {
                                // pass
                            }
                        }
                    }, reachablePingCts.Token).Forget();
                }
            }
            catch(Exception e)
            {
                Debug.LogError("PostmanClient :: connection error - " + e.Message);
            }
        }

        public void Close(bool asAsync = true)
        {
            Debug.Log("PostmanClient :: connection close");

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

            if(reachablePingCts != null)
                reachablePingCts.Cancel();
            reachablePingCts = null;
        }

        private IEnumerator ReconnectCoroutine()
        {
            float wait = 1.0f;

            while(true)
            {
                Debug.Log("PostmanClient :: try reconnect");

                Connect(true);

                yield return new WaitForSeconds(wait + UnityEngine.Random.Range(0.0f, 1.0f));

                if(isConnect && webSocket != null && webSocket.IsAlive)
                    break;

                if(exponentialBackoff)
                    wait *= 2;
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

#region ping
        public void Ping()
        {
            if(isConnect && webSocket != null)
                StartCoroutine(PingCoroutine());
        }

        private IEnumerator PingCoroutine()
        {
            webSocket.SendAsync(PostmanMassageData.BuildMessage(MessageType.PING), null);
            yield return null;
        }

        public bool InternalPing()
        {
            if(webSocket != null)
                return webSocket.Ping();
            else
                return false;
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
            {
                UniTask.RunOnThreadPool(() => {
                    webSocket.Send(PostmanMassageData.BuildMessage(MessageType.PUBLISH, json));
                }).Forget();
            }
            else
            {
                UniTask.RunOnThreadPool(() => {
                    webSocket.SendAsync(PostmanMassageData.BuildMessage(MessageType.PUBLISH, json), null);
                }).Forget();
            }

            yield return null;
        }
#endregion

#region store get
        public string StoreGet(string key)
        {
            ResultMessageData data = StoreGetAsData(key);
            return data.result;
        }

        public async UniTask<string> StoreGetAsync(string key)
        {
            ResultMessageData data = await StoreGetAsDataAsync(key);
            return data.result;
        }

        public ResultMessageData StoreGetAsData(string key)
        {
            string url = string.Format("{0}://{1}/postman/store?cmd=GET&key={2}",
                (useSSL ? "https" : "http"),
                host,
                Uri.EscapeDataString(key));
            if(secureToken != "")
                url += "&tkn=" + Uri.EscapeDataString(secureToken);

            UnityWebRequest request = UnityWebRequest.Get(url);
            request.SendWebRequest();

            while(!request.isDone);

            ResultMessageData responce;
#if UNITY_2020_1_OR_NEWER
            if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
#else
            if(request.isHttpError || request.isNetworkError)
#endif
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

        public async UniTask<ResultMessageData> StoreGetAsDataAsync(string key)
        {
            string url = string.Format("{0}://{1}/postman/store?cmd=GET&key={2}",
                (useSSL ? "https" : "http"),
                host,
                Uri.EscapeDataString(key));
            if(secureToken != "")
                url += "&tkn=" + Uri.EscapeDataString(secureToken);

            UnityWebRequest request = UnityWebRequest.Get(url);
            await request.SendWebRequest();

            ResultMessageData responce;
#if UNITY_2020_1_OR_NEWER
            if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
#else
            if(request.isHttpError || request.isNetworkError)
#endif
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
        public string StoreSet(string key, string val)
        {
            ResultMessageData data = StoreSetAsData(key, val);
            return data.result;
        }

        public async UniTask<string> StoreSetAsync(string key, string val)
        {
            ResultMessageData data = await StoreSetAsDataAsync(key, val);
            return data.result;
        }

        public ResultMessageData StoreSetAsData(string key, string val)
        {
            string url = string.Format("{0}://{1}/postman/store?cmd=SET&key={2}&val={3}",
                (useSSL ? "https" : "http"),
                host,
                Uri.EscapeDataString(key),
                Uri.EscapeDataString(val));
            if(secureToken != "")
                url += "&tkn=" + Uri.EscapeDataString(secureToken);

            UnityWebRequest request = UnityWebRequest.Get(url);
            request.SendWebRequest();

            while(!request.isDone);

            ResultMessageData responce = new ResultMessageData("", "");
#if UNITY_2020_1_OR_NEWER
            if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
#else
            if(request.isHttpError || request.isNetworkError)
#endif
                Debug.LogError("PostmanClient :: " + request.error);
            else
            {
                responce = JsonConvert.DeserializeObject<ResultMessageData>(request.downloadHandler.text);
                Debug.Log(string.Format("PostmanClient :: store set [ {0} : {1} ]", key, val));
            }

            request.Dispose();

            return responce;
        }

        public async UniTask<ResultMessageData> StoreSetAsDataAsync(string key, string val)
        {
            string url = string.Format("{0}://{1}/postman/store?cmd=SET&key={2}&val={3}",
                (useSSL ? "https" : "http"),
                host,
                Uri.EscapeDataString(key),
                Uri.EscapeDataString(val));
            if(secureToken != "")
                url += "&tkn=" + Uri.EscapeDataString(secureToken);

            UnityWebRequest request = UnityWebRequest.Get(url);
            await request.SendWebRequest();

            ResultMessageData responce = new ResultMessageData("", "");
#if UNITY_2020_1_OR_NEWER
            if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
#else
            if(request.isHttpError || request.isNetworkError)
#endif
                Debug.LogError("PostmanClient :: " + request.error);
            else
            {
                responce = JsonConvert.DeserializeObject<ResultMessageData>(request.downloadHandler.text);
                Debug.Log(string.Format("PostmanClient :: store set [ {0} : {1} ]", key, val));
            }

            request.Dispose();

            return responce;
        }
#endregion

#region store haskey
        public bool StoreHasKey(string key)
        {
            ResultMessageData data = StoreHasKeyAsData(key);

            bool b = false;
            bool.TryParse(data.result, out b);
            return b;
        }

        public ResultMessageData StoreHasKeyAsData(string key)
        {
            string url = string.Format("{0}://{1}/postman/store?cmd=HAS&key={2}",
                (useSSL ? "https" : "http"),
                host,
                Uri.EscapeDataString(key));
            if(secureToken != "")
                url += "&tkn=" + Uri.EscapeDataString(secureToken);

            UnityWebRequest request = UnityWebRequest.Get(url);
            request.SendWebRequest();

            while(!request.isDone);

            ResultMessageData responce;
#if UNITY_2020_1_OR_NEWER
            if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
#else
            if(request.isHttpError || request.isNetworkError)
#endif
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
            string url = string.Format("{0}://{1}/postman/store?cmd=DEL&key={2}",
                (useSSL ? "https" : "http"),
                host,
                Uri.EscapeDataString(key));
            if(secureToken != "")
                url += "&tkn=" + Uri.EscapeDataString(secureToken);

            UnityWebRequest request = UnityWebRequest.Get(url);
            request.SendWebRequest();

            while(!request.isDone);

#if UNITY_2020_1_OR_NEWER
            if(request.result == UnityWebRequest.Result.ProtocolError || request.result == UnityWebRequest.Result.ConnectionError)
#else
            if(request.isHttpError || request.isNetworkError)
#endif
                Debug.LogError("PostmanClient :: " + request.error);
            else
                Debug.Log(string.Format("PostmanClient :: store delete [ {0} ]", key));

            request.Dispose();
        }
#endregion

#endregion
    }
}
