import { useAuth } from '../lib/auth'

export default function Home() {
  const { user, logout, isAuthRequired } = useAuth()

  return (
    <div style={{ fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif', padding: '24px', maxWidth: '800px', margin: '0 auto' }}>
      <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '32px' }}>
        <h1 style={{ margin: 0, fontSize: '20px' }}>GenKitKraft</h1>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <a href="/settings" style={{ fontSize: '14px', color: '#666', textDecoration: 'none' }}>Settings</a>
          {isAuthRequired && user && (
            <>
              <span style={{ fontSize: '14px', color: '#666' }}>{user}</span>
              <button
                onClick={logout}
                style={{
                  padding: '6px 12px',
                  fontSize: '13px',
                  backgroundColor: 'transparent',
                  border: '1px solid #ddd',
                  borderRadius: '6px',
                  cursor: 'pointer',
                }}
              >
                Sign out
              </button>
            </>
          )}
        </div>
      </header>
      <main>
        <p style={{ color: '#666' }}>Agent configuration dashboard coming soon.</p>
      </main>
    </div>
  )
}
