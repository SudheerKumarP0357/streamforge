export default function DashboardLoading() {
  return (
    <div className="container" style={{ padding: '40px 0' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 32 }}>
        <div style={{ width: '200px', height: '36px', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '8px' }} className="animate-pulse" />
        <div style={{ width: '120px', height: '36px', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '8px' }} className="animate-pulse" />
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
        gap: 24,
      }}>
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} style={{
            backgroundColor: 'rgba(255,255,255,0.03)',
            border: '1px solid rgba(255,255,255,0.05)',
            borderRadius: 16,
            overflow: 'hidden',
          }} className="animate-pulse">
            <div style={{ width: '100%', aspectRatio: '16/9', backgroundColor: 'rgba(255,255,255,0.05)' }} />
            <div style={{ padding: 16 }}>
              <div style={{ height: '20px', width: '80%', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '4px', marginBottom: '12px' }} />
              <div style={{ height: '14px', width: '50%', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '4px', marginBottom: '16px' }} />
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div style={{ height: '24px', width: '64px', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '12px' }} />
                <div style={{ height: '14px', width: '40px', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: '4px' }} />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
