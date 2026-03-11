export default function WatchLoading() {
  return (
    <div style={{
      width: '100%',
      maxWidth: '1200px',
      margin: '0 auto',
      padding: '20px',
    }}>
      <div 
        className="animate-pulse"
        style={{
          width: '100%',
          aspectRatio: '16/9',
          backgroundColor: 'rgba(255,255,255,0.05)',
          borderRadius: '12px',
          marginBottom: '24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center'
        }}
      >
        {/* Spinner icon representation */}
        <svg className="animate-spin" width="48" height="48" fill="none" viewBox="0 0 24 24"
             style={{ color: 'rgba(255,255,255,0.2)' }}>
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      </div>

      <div className="animate-pulse" style={{ marginTop: '24px' }}>
        <div style={{ height: '32px', width: '60%', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '6px', marginBottom: '16px' }} />
        <div style={{ height: '16px', width: '20%', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '4px', marginBottom: '32px' }} />
        <div style={{ height: '16px', width: '100%', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '4px', marginBottom: '12px' }} />
        <div style={{ height: '16px', width: '80%', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '4px' }} />
      </div>
    </div>
  );
}
