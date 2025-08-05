import React, { useState } from 'react';
import { Web3Provider } from './contexts/Web3Context';
import Header from './components/Header';
import SearchBar from './components/SearchBar';
import NetworkStats from './components/NetworkStats';
import SearchResults from './components/SearchResults';
import PortfolioTracker from './components/PortfolioTracker';
import WalletSelector from './components/WalletSelector';
import { 
  apiService, 
  BlockResponse, 
  TransactionResponse, 
  BalanceResponse, 
  ContractInfoResponse 
} from './services/api';

interface SearchResult {
  type: 'block' | 'transaction' | 'address';
  data: BlockResponse | TransactionResponse | (BalanceResponse & { contractInfo?: ContractInfoResponse });
}

function App() {
  const [searchResult, setSearchResult] = useState<SearchResult | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'search' | 'portfolio'>('search');

  const handleSearch = async (query: string, type: 'block' | 'transaction' | 'address') => {
    setSearchLoading(true);
    setSearchError(null);
    setSearchResult(null);

    try {
      let data: any;

      switch (type) {
        case 'block':
          data = await apiService.getBlock(query);
          break;
        
        case 'transaction':
          data = await apiService.getTransaction(query);
          break;
        
        case 'address':
          // Fetch both balance and contract info for addresses
          const [balanceData, contractInfo] = await Promise.allSettled([
            apiService.getBalance(query),
            apiService.getContractInfo(query)
          ]);

          if (balanceData.status === 'fulfilled') {
            data = balanceData.value;
            if (contractInfo.status === 'fulfilled') {
              data.contractInfo = contractInfo.value;
            }
          } else {
            throw new Error('Failed to fetch address data');
          }
          break;
        
        default:
          throw new Error('Invalid search type');
      }

      setSearchResult({ type, data });
    } catch (error: any) {
      console.error('Search error:', error);
      setSearchError(
        error.response?.data?.message || 
        error.message || 
        'An error occurred while searching'
      );
    } finally {
      setSearchLoading(false);
    }
  };

  return (
    <Web3Provider>
      <div className="min-h-screen bg-secondary-50">
        <Header />
        <div className="flex justify-end p-4">
          <WalletSelector />
        </div>
        
        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Hero Section */}
          <div className="text-center mb-12">
            <h1 className="text-4xl font-bold text-gradient mb-4">
              Blockchain Analytics Dashboard
            </h1>
            <p className="text-xl text-secondary-600 max-w-3xl mx-auto">
              Explore and analyze blockchain data with powerful search capabilities. 
              Search for blocks, transactions, and addresses to get detailed insights.
            </p>
          </div>

          {/* Navigation Tabs */}
          <div className="flex justify-center mb-8">
            <div className="flex space-x-1 bg-white p-1 rounded-lg border border-secondary-200">
              <button
                onClick={() => setActiveTab('search')}
                className={`px-6 py-2 rounded-md font-medium transition-colors ${
                  activeTab === 'search'
                    ? 'bg-primary-600 text-white'
                    : 'text-secondary-600 hover:text-secondary-900'
                }`}
              >
                Blockchain Explorer
              </button>
              <button
                onClick={() => setActiveTab('portfolio')}
                className={`px-6 py-2 rounded-md font-medium transition-colors ${
                  activeTab === 'portfolio'
                    ? 'bg-primary-600 text-white'
                    : 'text-secondary-600 hover:text-secondary-900'
                }`}
              >
                Portfolio Tracker
              </button>
            </div>
          </div>

        {/* Content based on active tab */}
        {activeTab === 'search' ? (
          <>
            {/* Search Section */}
            <div className="mb-12">
              <SearchBar onSearch={handleSearch} loading={searchLoading} />
            </div>

        {/* Search Results */}
        {searchLoading && (
          <div className="mb-8">
            <SearchResults 
              type="block" 
              data={{} as BlockResponse} 
              loading={true} 
            />
          </div>
        )}

        {searchError && (
          <div className="mb-8">
            <div className="card">
              <div className="text-center py-8">
                <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 className="text-lg font-medium text-secondary-900 mb-2">
                  Search Failed
                </h3>
                <p className="text-secondary-500 mb-4">{searchError}</p>
                <button 
                  onClick={() => setSearchError(null)}
                  className="btn-secondary"
                >
                  Dismiss
                </button>
              </div>
            </div>
          </div>
        )}

        {searchResult && !searchLoading && (
          <div className="mb-8">
            <SearchResults 
              type={searchResult.type}
              data={searchResult.data}
            />
          </div>
        )}

            {/* Network Statistics */}
            <div className="mb-8">
              <NetworkStats />
            </div>

            {/* Getting Started Section */}
            {!searchResult && !searchLoading && !searchError && (
          <div className="card">
            <div className="text-center py-12">
              <div className="w-20 h-20 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-6">
                <svg className="w-10 h-10 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
              <h2 className="text-2xl font-bold text-secondary-900 mb-4">
                Start Exploring Blockchain Data
              </h2>
              <p className="text-secondary-600 mb-8 max-w-2xl mx-auto">
                Use the search bar above to explore blockchain data. You can search for:
              </p>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto">
                <div className="text-center">
                  <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mx-auto mb-3">
                    <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                  </div>
                  <h3 className="font-semibold text-secondary-900 mb-2">Blocks</h3>
                  <p className="text-sm text-secondary-600">
                    Search by block number or "latest" to get detailed block information
                  </p>
                </div>
                
                <div className="text-center">
                  <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mx-auto mb-3">
                    <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 4V2a1 1 0 011-1h8a1 1 0 011 1v2m-9 0h10m-9 0a2 2 0 00-2 2v14a2 2 0 002 2h8a2 2 0 002-2V6a2 2 0 00-2-2" />
                    </svg>
                  </div>
                  <h3 className="font-semibold text-secondary-900 mb-2">Transactions</h3>
                  <p className="text-sm text-secondary-600">
                    Enter a transaction hash to view transaction details and status
                  </p>
                </div>
                
                <div className="text-center">
                  <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mx-auto mb-3">
                    <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                    </svg>
                  </div>
                  <h3 className="font-semibold text-secondary-900 mb-2">Addresses</h3>
                  <p className="text-sm text-secondary-600">
                    Look up Ethereum addresses to view balances and contract information
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}
          </>
        ) : (
          /* Portfolio Tab */
          <PortfolioTracker />
        )}
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-secondary-200 mt-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <p className="text-secondary-500">
              © 2025 Kaia Analytics AI. Built with React, TypeScript, and Tailwind CSS.
            </p>
            <p className="text-sm text-secondary-400 mt-2">
              Powered by Ethereum blockchain data
            </p>
          </div>
        </div>
      </footer>
        </div>
      </Web3Provider>
    );
  }

export default App;
