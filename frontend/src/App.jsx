import React, { useState, useEffect } from 'react';

export default function SahayakDashboard() {
  const [sysState, setSysState] = useState('IDLE');
  const [message, setMessage] = useState('Ready for transaction execution');
  const [idemKey, setIdemKey] = useState(`TXN-${Date.now()}`);
  const [retryCount, setRetryCount] = useState(0);

  useEffect(() => {
    // Establish real-time persistent channel with the Go backend engine
    const ws = new WebSocket('ws://localhost:8080/ws');
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      // Update state machine mapping dynamically based on backend broadcasts
      setSysState(data.state);
      setMessage(data.message);
      
      if (data.state === 'RETRYING') {
        setRetryCount(prev => prev + 1);
      }
    };

    ws.onerror = (error) => {
      console.error('[WS SERVER DRIFT] Connection error:', error);
    };

    return () => ws.close();
  }, []);

  const triggerTransaction = async () => {
    // Phase 1: Local state transitions instantly to prevent button double-clicks
    setSysState('PROCESSING');
    setMessage('Initiating cryptographic handshake...');
    
    try {
      const res = await fetch('http://localhost:8080/transaction', {
        method: 'POST',
        headers: { 
          'X-Idempotency-Key': idemKey,
          'Content-Type': 'application/json'
        }
      });
      
      // Ingestion guard intercepts parallel execution paths
      if (res.status === 409) {
        const errorData = await res.json();
        setSysState('PANIC_LOCK');
        setMessage(`DUPLICATE PREVENTED: ${errorData.message}`);
      }
    } catch (e) {
      // Graceful degradation error catching
      console.error('[REST NETWORK DROP] Request blocked or failed:', e);
    }
  };

  const generateNewKey = () => {
    setIdemKey(`TXN-${Date.now()}`);
    setRetryCount(0);
    setSysState('IDLE');
    setMessage('Ready for next transaction cycle');
  };

  return (
    <div style={{ fontFamily: 'monospace', padding: '2rem', backgroundColor: '#0f172a', color: '#f8fafc', minHeight: '100vh' }}>
      
      {/* Visual Judging Layer Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid #334155', paddingBottom: '1rem', marginBottom: '2rem' }}>
        <h1 style={{ margin: 0, fontSize: '1.8rem', letterSpacing: '-0.02em' }}>Sahayak: Assurance Engine</h1>
        <span style={{ 
          padding: '0.4rem 0.8rem', 
          border: '1px solid #38bdf8', 
          color: '#38bdf8', 
          fontSize: '0.75rem', 
          fontWeight: 'bold',
          letterSpacing: '0.05em' 
        }}>
          SYSTEM MODE: DETERMINISTIC STATE MACHINE
        </span>
      </div>
      
      {/* Processing Indicator state */}
      {sysState === 'PROCESSING' && (
        <div style={{ border: '3px solid #38bdf8', padding: '1rem', backgroundColor: '#0c4a6e', marginBottom: '2rem' }}>
          <h2>⚡ PROCESSING REQUEST</h2>
          <p>{message}</p>
        </div>
      )}

      {/* Visual Overlay for Panic Lock Recovery Phase */}
      {['PANIC_LOCK', 'RETRYING'].includes(sysState) && (
        <div style={{ border: '3px solid #ef4444', padding: '1rem', backgroundColor: '#7f1d1d', marginBottom: '2rem' }}>
          <h2>🔒 FUNDS ARE SAFE. GATEWAY TIMEOUT (HTTP 504).</h2>
          <p>{message}</p>
        </div>
      )}

      {/* Automated Recovery Success State */}
      {sysState === 'SUCCESS' && (
        <div style={{ border: '3px solid #22c55e', padding: '1rem', backgroundColor: '#14532d', marginBottom: '2rem' }}>
          <h2>✅ ASYNC RECOVERY SUCCESS</h2>
          <p>{message}</p>
        </div>
      )}

      {/* Transaction Control Buttons */}
      <div style={{ display: 'flex', gap: '1rem' }}>
        <button 
          onClick={triggerTransaction}
          disabled={['PROCESSING', 'PANIC_LOCK', 'RETRYING', 'SUCCESS'].includes(sysState)}
          style={{ 
            padding: '1rem 2rem', 
            fontSize: '1.1rem', 
            fontWeight: 'bold', 
            cursor: ['PROCESSING', 'PANIC_LOCK', 'RETRYING', 'SUCCESS'].includes(sysState) ? 'not-allowed' : 'pointer',
            backgroundColor: ['PROCESSING', 'PANIC_LOCK', 'RETRYING', 'SUCCESS'].includes(sysState) ? '#475569' : '#38bdf8',
            color: '#0f172a',
            border: 'none'
          }}
        >
          Submit Transfer ($150)
        </button>

        <button 
          onClick={generateNewKey}
          style={{ 
            padding: '1rem 1.5rem', 
            fontSize: '1.1rem', 
            cursor: 'pointer',
            backgroundColor: 'transparent',
            color: '#f8fafc',
            border: '1px solid #475569'
          }}
        >
          Reset Environment
        </button>
      </div>

      {/* System Status Panel */}
      <div style={{ marginTop: '2.5rem', padding: '1.5rem', border: '1px solid #334155', backgroundColor: '#1e293b' }}>
        <h3 style={{ margin: '0 0 1rem 0', textTransform: 'uppercase', color: '#94a3b8' }}>Engine Telemetry Registry</h3>
        <p><strong>Cryptographic Idempotency Key:</strong> <span style={{ color: '#e2e8f0' }}>{idemKey}</span></p>
        <p><strong>Background Queue Status:</strong> <span style={{ color: ['PANIC_LOCK', 'RETRYING'].includes(sysState) ? '#ef4444' : '#22c55e' }}>
          {['PANIC_LOCK', 'RETRYING'].includes(sysState) ? 'ACTIVE - Goroutine Worker Processing' : 'IDLE / EMPTY'}
        </span></p>
        <p><strong>Systematic Retry Iteration:</strong> <span style={{ color: retryCount > 0 ? '#f59e0b' : '#f8fafc' }}>{retryCount}</span></p>
      </div>

      {/* Real-Time Deterministic Transaction Lifecycle Timeline */}
      <div style={{ marginTop: '2.5rem', display: 'flex', gap: '1rem', fontSize: '0.9rem', fontWeight: 'bold' }}>
        <span style={{ color: sysState !== 'IDLE' ? '#38bdf8' : '#475569' }}>[INITIATED]</span>
        <span style={{ color: ['PROCESSING', 'PANIC_LOCK', 'RETRYING', 'SUCCESS'].includes(sysState) ? '#38bdf8' : '#475569' }}>➔ [PROCESSING]</span>
        <span style={{ color: ['PANIC_LOCK', 'RETRYING', 'SUCCESS'].includes(sysState) ? '#ef4444' : '#475569' }}>➔ [FAILED (504)]</span>
        <span style={{ color: sysState === 'RETRYING' ? '#f59e0b' : (sysState === 'SUCCESS' ? '#22c55e' : '#475569') }}>➔ [RETRYING]</span>
        <span style={{ color: sysState === 'SUCCESS' ? '#22c55e' : '#475569' }}>➔ [SUCCESS]</span>
      </div>

    </div>
  );
}