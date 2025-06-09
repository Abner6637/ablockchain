import { useEffect, useState } from 'react';
import { jsonRpcCall } from './api/rpc';
import BlockCard from './components/BlockCard';
import BlockDetailModal from './components/BlockDetailModal';

function App() {
  const [chain, setChain] = useState([]);
  const [selectedBlock, setSelectedBlock] = useState(null);

  useEffect(() => {
    jsonRpcCall('getChain')
      .then(setChain)
      .catch(console.error);
  }, []);

  const handleBlockClick = async (index) => {
    try {
      const block = await jsonRpcCall('getBlockByIndex', [index]);
      setSelectedBlock(block);
    } catch (err) {
      alert('Failed to load block: ' + err.message);
    }
  };

  return (
    <div style={{ padding: '20px' }}>
      <h1>Blockchain Viewer</h1>
      <div style={{ display: 'flex', flexWrap: 'wrap' }}>
        {chain.map(block => (
          <BlockCard key={block.Header.Number} block={block} onClick={handleBlockClick} />
        ))}
      </div>
      {selectedBlock && (
        <BlockDetailModal block={selectedBlock} onClose={() => setSelectedBlock(null)} />
      )}
    </div>
  );
}

export default App;