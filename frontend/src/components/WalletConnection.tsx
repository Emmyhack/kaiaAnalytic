import React from 'react';
import { useAccount, useDisconnect, useEnsName } from 'wagmi';
import { Wallet, LogOut, Copy, ExternalLink } from 'lucide-react';
import { truncateAddress, copyToClipboard } from '../utils/format';

const WalletConnection: React.FC = () => {
  const { address, isConnected, chain } = useAccount();
  const { data: ensName } = useEnsName({ address });
  const { disconnect } = useDisconnect();

  const handleCopyAddress = async () => {
    if (address) {
      await copyToClipboard(address);
    }
  };

  const openBlockExplorer = () => {
    if (address && chain) {
      const explorerUrl = chain.blockExplorers?.default?.url;
      if (explorerUrl) {
        window.open(`${explorerUrl}/address/${address}`, '_blank');
      }
    }
  };

  const getChainName = (chainId?: number) => {
    switch (chainId) {
      case 1:
        return 'Ethereum';
      case 137:
        return 'Polygon';
      case 42161:
        return 'Arbitrum';
      case 11155111:
        return 'Sepolia';
      default:
        return 'Unknown';
    }
  };

  if (!isConnected || !address) {
    return (
      <div className="flex items-center">
        <button 
          onClick={() => (window as any).w3m?.open?.()}
          className="btn-primary"
        >
          <Wallet className="w-4 h-4 mr-2" />
          Connect Wallet
        </button>
      </div>
    );
  }

  return (
    <div className="flex items-center space-x-4">
      {/* Chain Indicator */}
      {chain && (
        <div className="flex items-center space-x-2 px-3 py-1 bg-secondary-100 rounded-full">
          <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
          <span className="text-sm font-medium text-secondary-700">
            {getChainName(chain.id)}
          </span>
        </div>
      )}

      {/* Wallet Info */}
      <div className="flex items-center space-x-3 px-4 py-2 bg-white rounded-lg border border-secondary-200 shadow-sm">
        <div className="flex items-center space-x-2">
          <Wallet className="w-4 h-4 text-primary-600" />
          <div className="flex flex-col">
            <span className="text-sm font-medium text-secondary-900">
              {ensName || truncateAddress(address)}
            </span>
            {ensName && (
              <span className="text-xs text-secondary-500">
                {truncateAddress(address)}
              </span>
            )}
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center space-x-1">
          <button
            onClick={handleCopyAddress}
            className="p-1 text-secondary-400 hover:text-secondary-600 transition-colors"
            title="Copy address"
          >
            <Copy className="w-4 h-4" />
          </button>
          
          <button
            onClick={openBlockExplorer}
            className="p-1 text-secondary-400 hover:text-secondary-600 transition-colors"
            title="View on explorer"
          >
            <ExternalLink className="w-4 h-4" />
          </button>

          <button
            onClick={() => disconnect()}
            className="p-1 text-secondary-400 hover:text-red-600 transition-colors"
            title="Disconnect wallet"
          >
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
};

export default WalletConnection;