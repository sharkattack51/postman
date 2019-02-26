using System.Collections;
using System.Collections.Generic;
using System;
using UnityEngine;
using WebSocketSharp;
using Newtonsoft.Json;

namespace Postman
{
	public class PostmanClient : MonoBehaviour
	{
		public string serverIp = "127.0.0.1:8800";
		public bool connectOnStart = true;
		public bool reconnectOnClose = false;
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

		void OnApplicationQuit()
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
				webSocket = new WebSocket(string.Format("ws://{0}/postman", ip));
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

		private IEnumerator PingCoroutine()
		{
			webSocket.SendAsync(PostmanMassageData.BuildMessage(MessageType.PING), null);
			yield return null;
		}

		public void Subscribe(string channel)
		{
			if(isConnect && webSocket != null && webSocket.IsAlive)
				StartCoroutine(SubscribeCoroutine(channel));
		}

		private IEnumerator SubscribeCoroutine(string channel)
		{
			SubscribeMessageData sub = new SubscribeMessageData(channel);
			string json = JsonConvert.SerializeObject(sub);
			webSocket.SendAsync(PostmanMassageData.BuildMessage(MessageType.SUBSCRIBE, json), null);
			yield return null;
		}

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
	}
}
