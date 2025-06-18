export default function BlockCard({ block, onClick }) {
  return (
    <div
      onClick={() => onClick(parseInt(block.Header.Number))}
      className="border border-gray-300 bg-white rounded-xl shadow-sm hover:shadow-lg p-4 cursor-pointer transition"
    >
      <h3 className="text-blue-600 font-bold text-lg">Block #{block.Header.Number}</h3>
      <p className="text-sm text-gray-500 truncate">Hash: {block.Hash}</p>
      <p className="text-sm text-gray-500">Tx Count: {block.Transactions.length}</p>
    </div>
  );
}