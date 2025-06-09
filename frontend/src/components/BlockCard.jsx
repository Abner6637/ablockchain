export default function BlockCard({ block, onClick }) {
  return (
    <div
      onClick={() => onClick(Number(block.Header.Number))}
      style={{
        border: '1px solid #ccc',
        padding: '10px',
        margin: '10px',
        cursor: 'pointer'
      }}
    >
      <p><strong>Number:</strong> {block.Header.Number}</p>
      <p><strong>Hash:</strong> {block.Hash.slice(0, 10)}...</p>
    </div>
  );
}