export default function BlockDetailModal({ block, onClose }) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center z-50">
      <div className="bg-white rounded-lg w-full max-w-2xl p-6 relative shadow-lg">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-gray-600 hover:text-red-500"
        >✕</button>
        <h2 className="text-2xl font-bold mb-4 text-blue-700">Block Details</h2>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div><strong>Number:</strong> {block.Header.Number}</div>
          <div><strong>Timestamp:</strong> {new Date(block.Header.Time * 1000).toLocaleString()}</div>
          <div className="col-span-2"><strong>Hash:</strong> {block.Hash}</div>
          <div className="col-span-2"><strong>Parent Hash:</strong> {block.Header.ParentHash}</div>
          <div><strong>Difficulty:</strong> {block.Header.Difficulty}</div>
          <div><strong>Nonce:</strong> {block.Header.Nonce}</div>
          <div><strong>Coinbase:</strong> {block.Header.Coinbase}</div>
        </div>
      </div>
    </div>
  );
}
