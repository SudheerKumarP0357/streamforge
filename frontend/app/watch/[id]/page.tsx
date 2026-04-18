import { cookies } from "next/headers";
import Link from "next/link";
import { Video } from "../../../lib/types";
import { serverApiUrl } from "../../../lib/config";
import { getUserIdFromToken } from "../../../lib/token";
import VideoPlayerWrapper from "../../../components/VideoPlayerWrapper";
import { logger } from "../../../lib/logger";

const BASE_URL = serverApiUrl;

export default async function WatchPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const renderStart = Date.now();
  const { id } = await params;
  const cookieStore = await cookies();
  const token = cookieStore.get("sf_access_token")?.value;
  const userId = token ? getUserIdFromToken(token) : null;

  logger.info('watch render start', { page: 'watch', video_id: id, user_id: userId });

  /* ── Auth guard ── */
  if (!token) {
    logger.warn('No access token, redirecting to login', { page: 'watch', video_id: id });
    logger.info('watch render complete', { page: 'watch', video_id: id, user_id: null, duration_ms: Date.now() - renderStart });
    return (
      <div style={{ paddingTop: "var(--nav-h)" }}>
        <div
          className="container-app"
          style={{ paddingTop: 80, textAlign: "center" }}
        >
          <p style={{ color: "var(--text-muted)", marginBottom: 16 }}>
            Please log in to watch videos.
          </p>
          <Link href="/login" className="btn btn-primary">
            Go to Login
          </Link>
        </div>
      </div>
    );
  }

  let video: Video | null = null;
  let streamUrl: string | null = null;
  let error: string | null = null;
  let sasToken: string | null = null;

  try {
    /* 1 — metadata */
    const metaPath = `/api/v1/videos/${id}`;
    const metaUrl = `${BASE_URL}${metaPath}`;

    logger.info('fetch start', { page: 'watch', action: 'GET', endpoint: metaPath, user_id: userId, video_id: id });
    const metaFetchStart = Date.now();

    const metaRes = await fetch(metaUrl, {
      headers: { Authorization: `Bearer ${token}` },
      next: { tags: [`video-${id}`], revalidate: 300 },
    });
    const metaDuration = Date.now() - metaFetchStart;

    logger.info('fetch result', {
      page: 'watch', action: 'GET', endpoint: metaPath,
      status: metaRes.status, success: metaRes.ok, duration_ms: metaDuration,
      user_id: userId, video_id: id,
    });

    if (!metaRes.ok) {
      error =
        metaRes.status === 404
          ? "Video not found."
          : "Failed to load video information.";
      logger.error(error, { page: 'watch', endpoint: metaPath, status: metaRes.status, video_id: id, user_id: userId });
      throw new Error(error);
    }
    video = await metaRes.json();

    /* 2 — stream URL */
    if (video?.status === "ready") {
      const streamPath = `/api/v1/videos/${id}/stream`;
      const streamEndpoint = `${BASE_URL}${streamPath}`;

      logger.info('fetch start', { page: 'watch', action: 'GET', endpoint: streamPath, user_id: userId, video_id: id });
      const streamFetchStart = Date.now();

      const streamRes = await fetch(
        streamEndpoint,
        {
          headers: { Authorization: `Bearer ${token}` },
          cache: 'no-store',   // ← SAS URLs expire — never cache this
        },
      );
      const streamDuration = Date.now() - streamFetchStart;

      logger.info('fetch result', {
        page: 'watch', action: 'GET', endpoint: streamPath,
        status: streamRes.status, success: streamRes.ok, duration_ms: streamDuration,
        user_id: userId, video_id: id,
      });

      if (streamRes.ok) {
        const streamData = await streamRes.json();
        streamUrl = streamData.master_playlist_url;
        sasToken = streamData.sas_token;
      } else {
        error = "Failed to retrieve stream URL.";
        logger.error(error, { page: 'watch', endpoint: streamPath, status: streamRes.status, video_id: id, user_id: userId });
      }
    } else {
      logger.info('Video not ready for streaming', { page: 'watch', video_id: id, status: video?.status, user_id: userId });
    }

    /* ── Error state ── */
    if (error || !video) {
      logger.info('watch render complete', { page: 'watch', video_id: id, user_id: userId, duration_ms: Date.now() - renderStart, outcome: 'error' });
      return (
        <div style={{ paddingTop: "var(--nav-h)" }}>
          <div
            className="container-app"
            style={{ maxWidth: 520, paddingTop: 60, paddingBottom: 60 }}
          >
            <div
              className="card"
              style={{
                padding: "40px 36px",
                textAlign: "center",
                position: "relative",
              }}
            >
              <div
                style={{
                  width: 60,
                  height: 60,
                  borderRadius: 14,
                  margin: "0 auto 20px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  background: "var(--error-bg)",
                  border: "1px solid var(--error-border)",
                }}
              >
                <svg
                  width="26"
                  height="26"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  style={{ color: "var(--error)", flexShrink: 0 }}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="1.5"
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                  />
                </svg>
              </div>
              <h2
                style={{
                  fontWeight: 600,
                  fontSize: "1.15rem",
                  color: "rgba(255,255,255,0.88)",
                  marginBottom: 8,
                }}
              >
                Error loading video
              </h2>
              <p
                style={{
                  color: "var(--text-muted)",
                  fontSize: "0.875rem",
                  marginBottom: 24,
                }}
              >
                {error || "Video not found"}
              </p>
              <Link href="/dashboard" className="btn btn-primary">
                Back to Dashboard
              </Link>
            </div>
          </div>
        </div>
      );
    }

    /* ── Not ready state ── */
    if (video.status !== "ready") {
      logger.info('watch render complete', { page: 'watch', video_id: id, user_id: userId, duration_ms: Date.now() - renderStart, outcome: video.status });
      const isProcessing = video.status === "processing";
      const isFailed = video.status === "failed";

      return (
        <div style={{ paddingTop: "var(--nav-h)" }}>
          <div
            className="container-app animate-fade-up"
            style={{ maxWidth: 520, paddingTop: 60, paddingBottom: 60 }}
          >
            <div
              className="card"
              style={{
                padding: "48px 40px",
                textAlign: "center",
                position: "relative",
                overflow: "hidden",
                border: isProcessing
                  ? "1px solid rgba(251,191,36,0.15)"
                  : undefined,
              }}
            >
              {/* Top accent bar */}
              <div
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  right: 0,
                  height: 1,
                  background: isProcessing
                    ? "linear-gradient(90deg,transparent,rgba(251,191,36,0.5) 50%,transparent)"
                    : "linear-gradient(90deg,transparent,rgba(245,200,66,0.4) 50%,transparent)",
                }}
              />

              {/* Icon */}
              <div
                style={{
                  width: 72,
                  height: 72,
                  borderRadius: 18,
                  margin: "0 auto 24px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  background: isProcessing
                    ? "rgba(251,191,36,0.08)"
                    : isFailed
                      ? "var(--error-bg)"
                      : "rgba(255,255,255,0.04)",
                  border: `1px solid ${isProcessing ? "rgba(251,191,36,0.22)" : isFailed ? "var(--error-border)" : "var(--border)"}`,
                }}
              >
                {isProcessing ? (
                  <svg
                    className="animate-spin"
                    width="32"
                    height="32"
                    fill="none"
                    viewBox="0 0 24 24"
                    style={{ color: "#fbbf24", flexShrink: 0 }}
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="3"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                    />
                  </svg>
                ) : (
                  <svg
                    width="32"
                    height="32"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    style={{
                      color: isFailed ? "var(--error)" : "var(--text-muted)",
                      flexShrink: 0,
                    }}
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="1.5"
                      d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                )}
              </div>

              <h1
                className="serif"
                style={{ fontSize: "1.75rem", color: "#fff", marginBottom: 10 }}
              >
                Video{" "}
                {isProcessing ? "Processing" : isFailed ? "Failed" : "Pending"}
              </h1>
              <p
                style={{
                  color: "var(--text-muted)",
                  fontSize: "0.9rem",
                  lineHeight: 1.7,
                  marginBottom: 28,
                  maxWidth: 340,
                  margin: "0 auto 28px",
                }}
              >
                {isProcessing
                  ? "Your video is being transcoded to adaptive bitrates. This usually takes a few minutes."
                  : isFailed
                    ? "Processing failed. Please try uploading the video again."
                    : "Your video is queued and will begin processing shortly."}
              </p>
              <Link href="/dashboard" className="btn btn-ghost">
                Return to Dashboard
              </Link>
            </div>
          </div>
        </div>
      );
    }

    /* ── Ready state ── */
    logger.info('watch render complete', { page: 'watch', video_id: id, user_id: userId, duration_ms: Date.now() - renderStart, outcome: 'ready' });

    return (
      <div style={{ paddingTop: "var(--nav-h)" }}>
        <div
          className="container-app"
          style={{ maxWidth: 900, paddingTop: 28, paddingBottom: 56 }}
        >
          {/* ── Breadcrumb ── */}
          <nav
            className="animate-fade-up"
            style={{
              display: "flex",
              alignItems: "center",
              gap: 6,
              fontSize: "0.8rem",
              marginBottom: 20,
            }}
          >
            <Link
              href="/dashboard"
              className="btn btn-ghost btn-sm"
              style={{
                display: "inline-flex",
                alignItems: "center",
                gap: 6,
              }}
            >
              <svg
                width="14"
                height="14"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                style={{ flexShrink: 0 }}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="1.5"
                  d="M10 19l-7-7m0 0l7-7m-7 7h18"
                />
              </svg>
              Dashboard
            </Link>
            <svg
              width="12"
              height="12"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              style={{ color: "rgba(255,255,255,0.2)", flexShrink: 0 }}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M9 5l7 7-7 7"
              />
            </svg>
            <span
              style={{
                color: "var(--text-muted)",
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
                maxWidth: 320,
              }}
            >
              {video.title}
            </span>
            <span
              className="badge badge-success"
              style={{ fontSize: "0.58rem", flexShrink: 0 }}
            >
              Ready
            </span>
          </nav>

          {/* ── Player ── */}
          <div
            className="animate-fade-up delay-75"
            style={{
              borderRadius: "var(--r-xl)",
              overflow: "hidden",
              border: "1px solid rgba(255,255,255,0.07)",
              boxShadow: "0 0 48px rgba(0,0,0,0.55)",
              marginBottom: 20,
              background: "#000",
            }}
          >
            {streamUrl ? (
              <VideoPlayerWrapper
                videoId={video.id}
                src={streamUrl}
                title={video.title}
                status={video.status}
                sasToken={sasToken ?? ''}
              />
            ) : (
              <div
                style={{
                  aspectRatio: "16/9",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  color: "var(--error)",
                  fontSize: "0.875rem",
                }}
              >
                Stream URL unavailable.
              </div>
            )}
          </div>

          {/* ── Video details ── */}
          <div
            className="card animate-fade-up delay-150"
            style={{ padding: "24px 28px" }}
          >
            <h1
              style={{
                fontSize: "1.3rem",
                fontWeight: 600,
                color: "rgba(255,255,255,0.92)",
                marginBottom: 6,
                lineHeight: 1.3,
              }}
            >
              {video.title}
            </h1>

            {video.description && (
              <p
                style={{
                  color: "var(--text-muted)",
                  fontSize: "0.9rem",
                  lineHeight: 1.7,
                  whiteSpace: "pre-wrap",
                  marginTop: 10,
                }}
              >
                {video.description}
              </p>
            )}

            {/* Meta row */}
            <div
              style={{
                display: "flex",
                flexWrap: "wrap",
                alignItems: "center",
                gap: 16,
                marginTop: 20,
                paddingTop: 16,
                borderTop: "1px solid rgba(255,255,255,0.06)",
                fontSize: "0.78rem",
                color: "var(--text-faint)",
              }}
            >
              <span style={{ display: "flex", alignItems: "center", gap: 6 }}>
                <svg
                  width="13"
                  height="13"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  style={{ color: "var(--accent)", flexShrink: 0 }}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                  />
                </svg>
                {new Date(video.created_at).toLocaleDateString("en-US", {
                  month: "long",
                  day: "numeric",
                  year: "numeric",
                })}
              </span>
              <span
                style={{
                  width: 3,
                  height: 3,
                  borderRadius: "50%",
                  background: "rgba(255,255,255,0.15)",
                  flexShrink: 0,
                }}
              />
              <span style={{ display: "flex", alignItems: "center", gap: 6 }}>
                <svg
                  width="13"
                  height="13"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  style={{ color: "var(--accent)", flexShrink: 0 }}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                {new Date(video.created_at).toLocaleTimeString("en-US", {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
            </div>
          </div>
        </div>
      </div>
    );
  } catch (caughtErr) {
    logger.error('watch render failed', caughtErr instanceof Error ? caughtErr : undefined, { page: 'watch', video_id: id, user_id: userId, duration_ms: Date.now() - renderStart });
    return (
      <div style={{ paddingTop: "var(--nav-h)" }}>
        <div
          className="container-app"
          style={{ maxWidth: 520, paddingTop: 60, paddingBottom: 60 }}
        >
          <div
            className="card"
            style={{
              padding: "40px 36px",
              textAlign: "center",
              position: "relative",
            }}
          >
            <div
              style={{
                width: 60,
                height: 60,
                borderRadius: 14,
                margin: "0 auto 20px",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                background: "var(--error-bg)",
                border: "1px solid var(--error-border)",
              }}
            >
              <svg
                width="26"
                height="26"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                style={{ color: "var(--error)", flexShrink: 0 }}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="1.5"
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
            </div>
            <h2
              style={{
                fontWeight: 600,
                fontSize: "1.15rem",
                color: "rgba(255,255,255,0.88)",
                marginBottom: 8,
              }}
            >
              Error loading video
            </h2>
            <p
              style={{
                color: "var(--text-muted)",
                fontSize: "0.875rem",
                marginBottom: 24,
              }}
            >
              Failed to load video information.
            </p>
            <Link href="/dashboard" className="btn btn-primary">
              Back to Dashboard
            </Link>
          </div>
        </div>
      </div>
    );
  }
}