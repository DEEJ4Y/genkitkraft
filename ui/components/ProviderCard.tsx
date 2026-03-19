import type { components } from '../lib/api/schema'

type Provider = components['schemas']['Models.ProviderResponse']

interface ProviderCardProps {
  providerType: string
  label: string
  provider: Provider | undefined
  onConfigure: () => void
  onDelete: () => void
}

export function ProviderCard({ providerType, label, provider, onConfigure, onDelete }: ProviderCardProps) {
  const configured = !!provider

  return (
    <div style={styles.card}>
      <div style={styles.header}>
        <div style={styles.titleRow}>
          <span style={{ ...styles.dot, backgroundColor: configured && provider.enabled ? '#4caf50' : '#ccc' }} />
          <h3 style={styles.name}>{label}</h3>
        </div>
        <span style={styles.badge}>
          {configured ? 'Configured' : 'Not configured'}
        </span>
      </div>

      {configured && (
        <div style={styles.details}>
          <div style={styles.detailRow}>
            <span style={styles.detailLabel}>API Key</span>
            <code style={styles.maskedKey}>{provider.apiKey}</code>
          </div>
          {provider.baseUrl && (
            <div style={styles.detailRow}>
              <span style={styles.detailLabel}>Base URL</span>
              <code style={styles.maskedKey}>{provider.baseUrl}</code>
            </div>
          )}
        </div>
      )}

      <div style={styles.actions}>
        <button onClick={onConfigure} style={styles.configureButton}>
          {configured ? 'Edit' : 'Configure'}
        </button>
        {configured && (
          <button onClick={onDelete} style={styles.deleteButton}>
            Delete
          </button>
        )}
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  card: {
    backgroundColor: '#fff',
    border: '1px solid #e0e0e0',
    borderRadius: '8px',
    padding: '20px',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '12px',
  },
  titleRow: {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
  },
  dot: {
    width: '8px',
    height: '8px',
    borderRadius: '50%',
    flexShrink: 0,
  },
  name: {
    margin: 0,
    fontSize: '16px',
    fontWeight: 600,
  },
  badge: {
    fontSize: '12px',
    color: '#666',
    padding: '2px 8px',
    backgroundColor: '#f5f5f5',
    borderRadius: '4px',
  },
  details: {
    marginBottom: '16px',
  },
  detailRow: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '4px 0',
  },
  detailLabel: {
    fontSize: '13px',
    color: '#666',
  },
  maskedKey: {
    fontSize: '13px',
    color: '#333',
    backgroundColor: '#f5f5f5',
    padding: '2px 6px',
    borderRadius: '4px',
  },
  actions: {
    display: 'flex',
    gap: '8px',
  },
  configureButton: {
    padding: '6px 16px',
    fontSize: '13px',
    fontWeight: 500,
    backgroundColor: '#111',
    color: '#fff',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
  },
  deleteButton: {
    padding: '6px 16px',
    fontSize: '13px',
    fontWeight: 500,
    backgroundColor: 'transparent',
    color: '#d32f2f',
    border: '1px solid #d32f2f',
    borderRadius: '6px',
    cursor: 'pointer',
  },
}
