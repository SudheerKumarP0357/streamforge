'use client';

import { useEffect, useRef, useState } from 'react';
import videojs from 'video.js';
import 'video.js/dist/video-js.css';
import type Player from 'video.js/dist/types/player';
import { logger } from '../lib/logger';

interface VideoPlayerProps {
  videoId: string;
  src: string;
  sasToken: string;       // ← already in interface, just needs to be used
  title?: string;
  status?: string;
}

function getCookie(name: string) {
  if (typeof document === 'undefined') return null;
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()?.split(';').shift() || null;
  return null;
}

export default function VideoPlayer({ videoId, src, sasToken, title, status }: VideoPlayerProps) {
  const videoRef = useRef<HTMLDivElement>(null);
  const playerRef = useRef<Player | null>(null);

  useEffect(() => {
    if (!videoRef.current) return;

    if (playerRef.current) {
      logger.info('Updating source on existing player', { component: 'VideoPlayer', video_id: videoId, src: src.substring(0, 80) });
      playerRef.current.src([{ src, type: 'application/x-mpegURL' }]);
      return;
    }

    logger.info('Initializing Video.js player', { component: 'VideoPlayer', video_id: videoId, src: src.substring(0, 80), hasSasToken: !!sasToken });

    const videoElement = document.createElement('video-js');
    videoElement.classList.add('vjs-big-play-centered');
    videoRef.current.appendChild(videoElement);

    const player = videojs(videoElement, {
      controls: true,
      autoplay: false,
      preload: 'auto',
      fluid: true,
      playbackRates: [0.5, 1, 1.25, 1.5, 2],
      sources: [{ src, type: 'application/x-mpegURL' }],
    });

    playerRef.current = player;

    // ── SAS token injection ──────────────────────────────────────────────────
    // Video.js VHS drops query strings when resolving relative .m3u8 and .ts
    // URLs. The onRequest hook fires before every network request and appends
    // the SAS token so Azure accepts the request.
    player.ready(() => {
      const tech = player.tech(true) as any;
      const vhs = tech?.vhs ?? tech?.hls; // 'vhs' in recent builds, 'hls' in older

      if (!vhs) {
        logger.warn('VHS tech not available — SAS injection skipped', { component: 'VideoPlayer', video_id: videoId });
        return;
      }

      logger.info('SAS token injection hook installed', { component: 'VideoPlayer', video_id: videoId });

      vhs.xhr.onRequest((options: any) => {
        const url: string = options.uri ?? '';
        if (!url || !sasToken) return options;

        // Only touch Azure blob storage requests
        if (!url.includes('.blob.core.windows.net')) return options;

        // Skip if the URL already carries a SAS signature
        if (url.includes('sig=')) return options;

        const sep = url.includes('?') ? '&' : '?';
        options.uri = `${url}${sep}${sasToken}`;
        return options;
      });
    });
    // ────────────────────────────────────────────────────────────────────────

    const trackEvent = (eventType: 'play' | 'pause' | 'end') => {
      const position = Math.floor(player.currentTime() ?? 0);
      const eventUrl = `/api/proxy/watch/${videoId}/event`;
      logger.info('Tracking watch event', { component: 'VideoPlayer', video_id: videoId, action: eventType, position, url: eventUrl });
      fetch(eventUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ event_type: eventType, position_seconds: position }),
      }).catch((err) => {
        logger.error('Failed to track watch event', { component: 'VideoPlayer', video_id: videoId, action: eventType });
      });
    };

    player.on('play', () => trackEvent('play'));
    player.on('pause', () => { if (!player.ended()) trackEvent('pause'); });
    player.on('ended', () => trackEvent('end'));

  }, [src, sasToken, videoId]);

  useEffect(() => {
    return () => {
      if (playerRef.current && !playerRef.current.isDisposed()) {
        playerRef.current.dispose();
        playerRef.current = null;
      }
    };
  }, []);

  return (
    <div className="w-full bg-black" style={{ aspectRatio: '16/9' }}>
      <div data-vjs-player style={{ height: '100%' }}>
        <div ref={videoRef} style={{ height: '100%' }} />
      </div>
    </div>
  );
}