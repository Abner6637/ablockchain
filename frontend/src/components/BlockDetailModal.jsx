export default function BlockDetailModal({ block, onClose }) {
  return (
    <div
      style={{
        position: 'fixed',
        top: '20%',
        left: '20%',
        width: '60%',
        backgroundColor: 'white',
        border: '1px solid #333',
        padding: '20px',
        zIndex: 1000
      }}
    >
      <h2>Block Detail</h2>
      <p><strong>Number:</strong> {block.Header.Number}</p>
      <p><strong>Timestamp:</strong> {block.Header.Time}</p>
      <p><strong>Hash:</strong> {block.Hash}</p>
      <p><strong>Parent Hash:</strong> {block.Header.ParentHash}</p>
      <button onClick={onClose}>Close</button>
    </div>
  );
}