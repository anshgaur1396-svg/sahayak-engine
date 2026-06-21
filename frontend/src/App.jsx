import React, { useState, useEffect } from 'react';

export default function SahayakDashboard() {
  const [sysState, setSysState] = useState('IDLE');
  const [message, setMessage] = useState('Ready for transaction');
  const [idemKey, setIdemKey] = useState(`TXN-${Date.now()}`);
  const [retryCount, setRetryCount] = useState(0);

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws');
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setSysState(data.state);
      setMessage(data.message);
      if (data.state === 'RETRYING') setRetryCount(prev => prev + 1);
    };
    return () => ws.close();
  }, []);

  const triggerTransaction = async () => {
    setSysState('PROCESSING');
    setMessage('Initiating cryptographic lock...');
    try {
      const res = await fetch('http://localhost:8080/transaction', {
        method: 'POST',
        headers: { 'X-Idempotency-Key': idemKey }
      });
      if (res.status === 409) {
        setSysState('PANIC_LOCK');
        setMessage('DUPLICATE PREVENTED: Transaction already secured.');
      }
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div style={{ fontFamily: 'monospace', padding: '2rem', backgroundColor: '#0f172a', color: '#f8fafc', height: '100vh' }}>
      <h1>Sahayak: Assurance Engine</h1>
      
      {/* Visual Overlay for Panic Lock */}
      {['PANIC_LOCK', 'RETRYING'].includes(sysState) && (
        <div style={{ border: '3px solid #ef4444', padding: '1rem', backgroundColor: '#7f1d1d', marginBottom: '2rem' }}>
          <h2>🔒 FUNDS ARE SAFE. NETWORK PAUSED.</h2>
          <p>{message}</p>
        </div>
      )}

      {/* Success State */}
      {sysState === 'SUCCESS' && (
        <div style={{ border: '3px solid #22c55e', padding: '1rem', backgroundColor: '#14532d', marginBottom: '2rem' }}>
          <h2>✅ RECOVERY SUCCESS</h2>
          <p>{message}</p>
        </div>
      )}

      <button 
        onClick={triggerTransaction}
        disabled={['PROCESSING', 'PANIC_LOCK', 'RETRYING'].includes(sysState)}
        style={{ padding: '1rem 2rem', fontSize: '1.2rem', cursor: 'pointer' }}
      >
        Submit Transfer ($150)
      </button>

      {/* System Status Panel */}
      <div style={{ marginTop: '2rem', padding: '1rem', border: '1px solid #475569' }}>
        <h3>System Status Panel</h3>
        <p><strong>Idempotency Key:</strong> {idemKey}</p>
        <p><strong>Queue Status:</strong> {['PANIC_LOCK', 'RETRYING'].includes(sysState) ? 'ACTIVE - Worker Processing' : 'EMPTY'}</p>
        <p><strong>Retry Count:</strong> {retryCount}</p>
      </div>

      {/* Transaction Timeline */}
      <div style={{ marginTop: '2rem', display: 'flex', gap: '1rem', opacity: 0.8 }}>
        <span style={{ color: sysState !== 'IDLE' ? '#38bdf8' : 'gray' }}>INITIATED →</span>
        <span style={{ color: ['PANIC_LOCK', 'RETRYING'].includes(sysState) ? '#ef4444' : 'gray' }}>FAILED (504) →</span>
        <span style={{ color: sysState === 'RETRYING' ? '#f59e0b' : 'gray' }}>RETRYING →</span>
        <span style={{ color: sysState === 'SUCCESS' ? '#22c55e' : 'gray' }}>SUCCESS</span>
      </div>
    </div>
  );
}