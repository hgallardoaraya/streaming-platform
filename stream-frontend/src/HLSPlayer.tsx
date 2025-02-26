import { useEffect, useRef } from 'react';
import Hls from 'hls.js';

// Componente correspondiente al reproductor de video a partir del servidor de distribución HLS.
// Por ahora el src viene hardcodeado, una mejora a futuro es que se elija el stream de manera dinámica.
const HLSPlayer = ({ src }: { src: string }) => {
  const videoRef = useRef<HTMLVideoElement>(null); 

  useEffect(() => {
    const video = videoRef.current;

    if (video) {
      if (Hls.isSupported()) {
        const hls = new Hls();
        hls.loadSource(src);
        hls.attachMedia(video);

        hls.on(Hls.Events.MANIFEST_PARSED, () => {
          video.play();
        });

        return () => hls.destroy();
      } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = src;
        video.addEventListener('loadedmetadata', () => {
          video.play();
        });
      }
    }
  }, [src]);

  return <video ref={videoRef} controls style={{ width: '100%' }} />;
};

export default HLSPlayer;
