import React, { useState, useEffect, useRef } from "react";


type DBMessage =   {
  "_id": string,
  "message"?: string,
  "streamKey"?: string,
  "user"?: string,
  "offset": number,
  "partition": number,
  "timestamp": string
}

type Message = {
  user: string,
  text: string,
}

const Chat: React.FC<{username: string}> = ({ username }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [message, setMessage] = useState<string>(""); 
  const socket = useRef<WebSocket | null>(null); 
  const chatRef = useRef<HTMLDivElement | null>(null); 

  useEffect(() => {
    //Obtiene mensajes anteriores de la BD
    getMessagesWrapper()
    
    // Conexión al WebSocket
    socket.current = new WebSocket("ws://localhost:9090/ws");

    // Maneja mensajes recibidos
    socket.current.onmessage = (event: MessageEvent) => {
      console.log(event.data);
      const socketMessage = JSON.parse(event.data);
      const newMessage:Message = {user: socketMessage.user, text: socketMessage.message}
      setMessages((prevMessages) => [...prevMessages, newMessage]);
    };

    // Cierre conexión WebSocket
    return () => {
      if (socket.current) {
        socket.current.close();
      }
    };
  }, []);

  useEffect(() => {
    if (chatRef.current) {
      chatRef.current.scrollTop = chatRef.current.scrollHeight;
    }
  }, [messages]);

  const sendMessage = () => {
    if (message.trim() && socket.current) {
      const messageToSend = {
        message,
        user: username,
      };
      socket.current.send(JSON.stringify(messageToSend));
      setMessage("");
    }

  };

  const getMessagesWrapper = async () => {
    await getMessages()
  }

  const getMessages = async () => {
    try {
      const response = await fetch("http://localhost:9091/messages", {method: "GET"});
      const msgs = await response.json()
      setMessages(msgs.map((msg:DBMessage) => ({user: msg.user, text: msg.message})))
    } catch(error: any) {
      console.log("Error al buscar mensajes");
    }
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setMessage(e.target.value);
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      sendMessage();
    }
  };

  return (
    <div style={styles.container}>
      <h1>Chat</h1>
      <div ref={chatRef} style={styles.chat}>
        {messages.map((msg, index) => (
          <div key={index} style={{
            margin: "5px 0",
            padding: "5px",
            borderRadius: "5px",
            backgroundColor: username == msg.user ? "#ffffff" : "#f0f0f0",            
            textAlign: username == msg.user ? "right" : "left",
          }}>
            {username === msg.user ? "Yo" : msg.user}: {msg.text}
          </div>
        ))}
      </div>
      <div style={styles.inputContainer}>
        <input
          type="text"
          value={message}
          onChange={handleInputChange}
          onKeyPress={handleKeyPress}
          placeholder="Type your message..."
          style={styles.input}
        />
        <button onClick={sendMessage} style={styles.button}>
          Send
        </button>
      </div>
    </div>
  );
};

const styles = {
  container: {
    fontFamily: "Arial, sans-serif",
    padding: "20px",
    textAlign: "center",
  },
  chat: {
    border: "1px solid #ccc",
    borderRadius: "5px",
    height: "300px",
    overflowY: "auto",
    padding: "10px",
    marginBottom: "10px",
    textAlign: "left",
  },
  inputContainer: {
    display: "flex",
    justifyContent: "space-between",
  },
  input: {
    flex: 1,
    padding: "10px",
    marginRight: "10px",
    borderRadius: "5px",
    border: "1px solid #ccc",
  },
  button: {
    padding: "10px 20px",
    borderRadius: "5px",
    border: "none",
    backgroundColor: "#007BFF",
    color: "#fff",
    cursor: "pointer",
  },
} as const;

export default Chat;
