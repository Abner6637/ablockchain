export async function jsonRpcCall(method, params = [], id = 1) {
  const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:26666';

  
  const response = await fetch(`${API_URL}/rpc`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      jsonrpc: '2.0',
      method,
      params,
      id,
    }),
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  if (data.error) throw new Error(data.error.message || 'RPC error');
  return data.result;
}