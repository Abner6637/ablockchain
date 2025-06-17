import { useEffect, useState } from 'react';
import { jsonRpcCall } from './api/rpc';
import BlockCard from './components/BlockCard';
import BlockDetailModal from './components/BlockDetailModal';
import Pagination from './components/Pagination';

export default function App() {
  const [chain, setChain] = useState([]);
  const [selectedBlock, setSelectedBlock] = useState(null);
  const [page, setPage] = useState(0);
  const blocksPerPage = 10;

  useEffect(() => {
    jsonRpcCall('getChain')
      .then(setChain)
      .catch(console.error);
  }, []);

  const totalPages = Math.ceil(chain.length / blocksPerPage);
  const pagedChain = chain.slice(page * blocksPerPage, (page + 1) * blocksPerPage);

  const handleBlockClick = async (index) => {
    try {
      const block = await jsonRpcCall('getBlockByIndex', [index]);
      setSelectedBlock(block);
    } catch (err) {
      alert('Failed to load block: ' + err.message);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <h1 className="text-3xl font-bold text-center text-blue-800 mb-8">
        Blockchain Explorer
      </h1>

      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
        {pagedChain.map(block => (
          <BlockCard key={block.Header.Number} block={block} onClick={handleBlockClick} />
        ))}
      </div>

      <Pagination
        page={page}
        totalPages={totalPages}
        onPrev={() => setPage(p => Math.max(0, p - 1))}
        onNext={() => setPage(p => Math.min(totalPages - 1, p + 1))}
      />

      {selectedBlock && (
        <BlockDetailModal block={selectedBlock} onClose={() => setSelectedBlock(null)} />
      )}
    </div>
  );
}