import React, { useState } from 'react';
import { Search, X } from 'lucide-react';
import { isValidAddress, isValidTxHash } from '../utils/format';

interface SearchBarProps {
  onSearch: (query: string, type: 'block' | 'transaction' | 'address') => void;
  loading?: boolean;
}

const SearchBar: React.FC<SearchBarProps> = ({ onSearch, loading = false }) => {
  const [query, setQuery] = useState('');
  const [error, setError] = useState('');

  const detectSearchType = (input: string): 'block' | 'transaction' | 'address' | null => {
    const trimmed = input.trim();
    
    // Check if it's a block number
    if (/^\d+$/.test(trimmed) || trimmed.toLowerCase() === 'latest') {
      return 'block';
    }
    
    // Check if it's a transaction hash
    if (isValidTxHash(trimmed)) {
      return 'transaction';
    }
    
    // Check if it's an address
    if (isValidAddress(trimmed)) {
      return 'address';
    }
    
    return null;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!query.trim()) {
      setError('Please enter a search query');
      return;
    }

    const type = detectSearchType(query);
    
    if (!type) {
      setError('Invalid input. Please enter a block number, transaction hash, or Ethereum address.');
      return;
    }

    setError('');
    onSearch(query.trim(), type);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value);
    if (error) setError('');
  };

  const clearSearch = () => {
    setQuery('');
    setError('');
  };

  const getPlaceholder = () => {
    return 'Search by block number, transaction hash, or address...';
  };

  const getSearchHint = () => {
    if (!query.trim()) return null;
    
    const type = detectSearchType(query);
    switch (type) {
      case 'block':
        return 'Block number detected';
      case 'transaction':
        return 'Transaction hash detected';
      case 'address':
        return 'Ethereum address detected';
      default:
        return 'Invalid format';
    }
  };

  return (
    <div className="w-full max-w-2xl mx-auto">
      <form onSubmit={handleSubmit} className="relative">
        <div className="relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Search className="h-5 w-5 text-secondary-400" />
          </div>
          
          <input
            type="text"
            value={query}
            onChange={handleInputChange}
            placeholder={getPlaceholder()}
            className={`input-field pl-10 pr-20 py-3 text-lg ${
              error ? 'border-red-300 focus:border-red-500 focus:ring-red-500' : ''
            }`}
            disabled={loading}
          />
          
          <div className="absolute inset-y-0 right-0 flex items-center">
            {query && (
              <button
                type="button"
                onClick={clearSearch}
                className="p-2 text-secondary-400 hover:text-secondary-600 transition-colors"
                disabled={loading}
              >
                <X className="h-4 w-4" />
              </button>
            )}
            
            <button
              type="submit"
              disabled={loading || !query.trim()}
              className="btn-primary ml-2 mr-2 px-4 py-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? (
                <div className="loading-spinner w-4 h-4" />
              ) : (
                'Search'
              )}
            </button>
          </div>
        </div>
      </form>

      {/* Search hint */}
      {query.trim() && (
        <div className="mt-2 flex items-center justify-between">
          <div className="text-sm text-secondary-500">
            <span className="font-medium">{getSearchHint()}</span>
            {detectSearchType(query) && (
              <span className="ml-2 px-2 py-1 bg-primary-50 text-primary-700 rounded-full text-xs">
                {detectSearchType(query)}
              </span>
            )}
          </div>
        </div>
      )}

      {/* Error message */}
      {error && (
        <div className="mt-2 text-sm text-red-600 bg-red-50 border border-red-200 rounded-md p-3">
          {error}
        </div>
      )}

      {/* Search examples */}
      <div className="mt-4 text-xs text-secondary-500">
        <div className="flex flex-wrap gap-4">
          <span><strong>Examples:</strong></span>
          <span>Block: <code className="bg-secondary-100 px-1 rounded">latest</code> or <code className="bg-secondary-100 px-1 rounded">18500000</code></span>
          <span>Address: <code className="bg-secondary-100 px-1 rounded">0x742d35Cc6aB...</code></span>
          <span>Transaction: <code className="bg-secondary-100 px-1 rounded">0x1234abcd...</code></span>
        </div>
      </div>
    </div>
  );
};

export default SearchBar;