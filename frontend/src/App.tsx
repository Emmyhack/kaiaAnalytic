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

// Error Boundary Component
class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { hasError: boolean; error?: Error }
> {
  constructor(props: { children: React.ReactNode }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: '20px', textAlign: 'center' }}>
          <h1>Something went wrong.</h1>
          <p>Error: {this.state.error?.message}</p>
          <button onClick={() => this.setState({ hasError: false })}>
            Try again
          </button>
        </div>
      );
    }

    return this.props.children;
  }
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
            
            // Add contract info if available
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

      setSearchResult({
        type,
        data
      });
    } catch (error) {
      console.error('Search error:', error);
      setSearchError(error instanceof Error ? error.message : 'An error occurred during search');
    } finally {
      setSearchLoading(false);
    }
  };

  // Simple fallback render for testing
  const renderSimpleVersion = () => (
    <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
      <h1>KaiaAnalyticsAI</h1>
      <p>Loading analytics platform...</p>
      <div style={{ marginTop: '20px', padding: '10px', backgroundColor: '#f0f0f0', borderRadius: '5px' }}>
        <p><strong>Status:</strong> App is running</p>
        <p><strong>Environment:</strong> {import.meta.env.MODE || 'unknown'}</p>
        <p><strong>Network ID:</strong> {import.meta.env.VITE_REACT_APP_NETWORK_ID || 'not set'}</p>
      </div>
    </div>
  );

  // For debugging, temporarily return the simple version
  return renderSimpleVersion();

  /*
  return (
    <ErrorBoundary>
      <Web3Provider>
        <div className="min-h-screen bg-secondary-50">
          <Header />
          
          <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 pb-8">
            <div className="text-center mb-8">
              <h1 className="text-4xl md:text-6xl font-bold text-secondary-900 mb-4">
                Kaia<span className="text-primary-600">Analytics</span>AI
              </h1>
              <p className="text-xl text-secondary-600 max-w-3xl mx-auto">
                Advanced blockchain analytics platform powered by AI for comprehensive
                transaction analysis, smart contract insights, and network intelligence.
              </p>
            </div>

            <div className="mb-8">
              <div className="flex justify-center space-x-4 mb-6">
                <button
                  onClick={() => setActiveTab('search')}
                  className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                    activeTab === 'search'
                      ? 'bg-primary-600 text-white shadow-lg'
                      : 'bg-white text-secondary-700 hover:bg-secondary-50'
                  }`}
                >
                  Analytics Search
                </button>
                <button
                  onClick={() => setActiveTab('portfolio')}
                  className={`px-6 py-3 rounded-lg font-semibold transition-all ${
                    activeTab === 'portfolio'
                      ? 'bg-primary-600 text-white shadow-lg'
                      : 'bg-white text-secondary-700 hover:bg-secondary-50'
                  }`}
                >
                  Portfolio Tracker
                </button>
              </div>

              {activeTab === 'search' ? (
                <div className="space-y-6">
                  <SearchBar onSearch={handleSearch} loading={searchLoading} />
                  
                  {searchError && (
                    <div className="max-w-2xl mx-auto p-4 bg-red-50 border border-red-200 rounded-lg">
                      <p className="text-red-800">{searchError}</p>
                    </div>
                  )}
                  
                  {searchResult && (
                    <SearchResults result={searchResult} />
                  )}
                  
                  <NetworkStats />
                </div>
              ) : (
                <div className="space-y-6">
                  <WalletSelector />
                  <PortfolioTracker />
                </div>
              )}
            </div>
          </main>
        </div>
      </Web3Provider>
    </ErrorBoundary>
  );
  */
}

export default App;
