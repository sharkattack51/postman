using System.Collections;
using System.Collections.Generic;
using UnityEngine;

namespace Postman
{
	public enum MessageType
	{
		PING = 0,
		SUBSCRIBE,
		UNSUBSCRIBE,
		PUBLISH
	}

	
	public class PostmanMassageData
	{
		public const string ProtocolMessageTag = "message ";
		public const string PingReturnString = "\"pong\"";

		public static string BuildMessage(MessageType type, string body = "{}")
		{
			string msg = "";

			switch(type)
			{
				case MessageType.PING: msg = string.Format("ping {0}", body); break;
				case MessageType.SUBSCRIBE: msg = string.Format("subscribe {0}", body); break;
				case MessageType.UNSUBSCRIBE: msg = string.Format("unsubscribe {0}", body); break;
				case MessageType.PUBLISH: msg = string.Format("publish {0}", body); break;
			}

			return msg;
		}
	}


	public class SubscribeMessageData : PostmanMassageData
	{
		public string channel;

		public SubscribeMessageData(string channel)
		{
			this.channel = channel;
		}
	}

	public class PublishMessageData : PostmanMassageData
	{
		public string channel;
		public string message;
		public string tag;
		public string extention;

		public PublishMessageData(string channel, string message, string tag = "", string extention = "")
		{
			this.channel = channel;
			this.message = message;
			this.tag = tag;
			this.extention = extention;
		}
	}

	public class ResultMessageData : PostmanMassageData
	{
		public string result;
		public string error;

		public ResultMessageData(string result, string error)
		{
			this.result = result;
			this.error = error;
		}
	}
}
