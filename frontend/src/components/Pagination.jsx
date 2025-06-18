export default function Pagination({ page, totalPages, onPrev, onNext }) {
  return (
    <div className="flex justify-center mt-6 space-x-4">
      <button
        onClick={onPrev}
        disabled={page === 0}
        className="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-50"
      >
        Prev
      </button>
      <span className="self-center font-medium text-gray-700">
        Page {page + 1} of {totalPages}
      </span>
      <button
        onClick={onNext}
        disabled={page + 1 >= totalPages}
        className="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-50"
      >
        Next
      </button>
    </div>
  );
}