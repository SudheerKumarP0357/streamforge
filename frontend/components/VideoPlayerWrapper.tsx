'use client';

import dynamic from 'next/dynamic';

const VideoPlayer = dynamic(
  () => import('./VideoPlayer'),
  {
    ssr: false,
    loading: () => (
      <div className="w-full max-w-4xl mx-auto mt-6 h-64 bg-gray-900 rounded-lg animate-pulse flex items-center justify-center">
        <span className="text-gray-500 text-sm">Loading player...</span>
      </div>
    )
  }
);

export default function VideoPlayerWrapper({
  videoId,
  src,
  title,
  status,
  sasToken,
}: {
  videoId: string;
  src: string;
  title: string;
  status: string;
  sasToken: string;
}) {
  return (
    <VideoPlayer
      videoId={videoId}
      src={src}
      title={title}
      status={status}
      sasToken={sasToken}
    />
  );
}
