import { useEffect, useState } from 'react';
import Chat from './Chat';
import HLSPlayer from './HLSPlayer';

const App = () => {
  const [username, setUsername] = useState("");
  const [streamerUsername, setStreamerUsername] = useState("");
  const [streams, setStreams] = useState([]);

  const getStreams = async () => {
    console.log("get streams");
    const resp = await fetch("http://localhost:9092/api/streams");
    const json = await resp.json();
    if(json) {
      setStreams(json);
    }
  }

  useEffect(() => {
    getStreams();
  }, [])

  return (
    <div>      
      <div>
        <label>
          Nombre de usuario: <input value={username} onChange={(e) => setUsername(e.target.value)}></input>          
        </label>
        <label>
          Nombre de usuario del streamer: <input value={streamerUsername} onChange={(e) => setStreamerUsername(e.target.value)}></input>          
        </label>
      </div>
      <div>
        <ul>        
          {streams.map((m:any) => (
            <li>{m.stream_name} {m.username} {m.status} {m.created_at}</li>
          ))} 
        </ul>       
      </div>
      <div style={{display: "flex"}}>             
        <div style={{width: "50%"}}>
          <h1>Video</h1>
          <HLSPlayer src={`http://localhost:3000/hls/${streamerUsername}/${streamerUsername}.m3u8`} />
        </div>
        <div style={{width: "50%"}}>
          <Chat username={username}/>
        </div>
      </div>
    </div>
  );
};

export default App;
