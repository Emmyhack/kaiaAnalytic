import React from 'react';
import { 
  Copy, 
  Clock, 
  Hash, 
  User, 
  Blocks,
  Fuel,
  DollarSign,
  CheckCircle,
  XCircle
} from 'lucide-react';
import { 
  BlockResponse, 
  TransactionResponse, 
  BalanceResponse, 
  ContractInfoResponse 
} from '../services/api';
import {
  truncateAddress,
  formatEther,
  formatGas,
  formatTimestamp,
  formatRelativeTime,
  formatBlockNumber,
  formatBytes,
  copyToClipboard
} from '../utils/format';

interface SearchResultsProps {
  type: 'block' | 'transaction' | 'address';
  data: BlockResponse | TransactionResponse | (BalanceResponse & { contractInfo?: ContractInfoResponse });
  loading?: boolean;
}

const SearchResults: React.FC<SearchResultsProps> = ({ type, data, loading }) => {
  const handleCopy = async (text: string) => {
    const success = await copyToClipboard(text);
    if (success) {
      // You could add a toast notification here
      console.log('Copied to clipboard');
    }
  };

  const CopyButton: React.FC<{ text: string; label?: string }> = ({ text, label }) => (
    <button
      onClick={() => handleCopy(text)}
      className="inline-flex items-center text-sm text-secondary-500 hover:text-secondary-700 transition-colors"
      title={`Copy ${label || 'value'}`}
    >
      <Copy className="w-4 h-4" />
    </button>
  );

  const InfoRow: React.FC<{ 
    label: string; 
    value: React.ReactNode; 
    copyable?: string;
    icon?: React.ReactNode;
  }> = ({ label, value, copyable, icon }) => (
    <div className="flex items-center justify-between py-3 border-b border-secondary-100 last:border-b-0">
      <div className="flex items-center space-x-2">
        {icon && <div className="text-secondary-400">{icon}</div>}
        <span className="text-sm font-medium text-secondary-600">{label}</span>
      </div>
      <div className="flex items-center space-x-2">
        <span className="text-sm text-secondary-900 font-mono break-all">{value}</span>
        {copyable && <CopyButton text={copyable} label={label} />}
      </div>
    </div>
  );

  if (loading) {
    return (
      <div className="card animate-pulse">
        <div className="space-y-4">
          <div className="h-6 bg-secondary-200 rounded w-1/4"></div>
          {[...Array(6)].map((_, i) => (
            <div key={i} className="flex justify-between items-center">
              <div className="h-4 bg-secondary-200 rounded w-1/3"></div>
              <div className="h-4 bg-secondary-200 rounded w-1/2"></div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  if (type === 'block') {
    const blockData = data as BlockResponse;
    return (
      <div className="card">
        <div className="card-header">
          <div className="flex items-center space-x-2">
            <Blocks className="w-5 h-5 text-primary-600" />
            <h3 className="text-lg font-semibold text-secondary-900">
              Block #{formatBlockNumber(blockData.number)}
            </h3>
          </div>
        </div>
        
        <div className="space-y-0">
          <InfoRow
            label="Block Hash"
            value={blockData.hash}
            copyable={blockData.hash}
            icon={<Hash className="w-4 h-4" />}
          />
          <InfoRow
            label="Parent Hash"
            value={truncateAddress(blockData.parent_hash, 8)}
            copyable={blockData.parent_hash}
            icon={<Hash className="w-4 h-4" />}
          />
          <InfoRow
            label="Timestamp"
            value={`${formatTimestamp(blockData.timestamp)} (${formatRelativeTime(blockData.timestamp)})`}
            icon={<Clock className="w-4 h-4" />}
          />
          <InfoRow
            label="Transactions"
            value={blockData.transaction_count.toLocaleString()}
            icon={<User className="w-4 h-4" />}
          />
          <InfoRow
            label="Gas Used"
            value={`${formatGas(blockData.gas_used)} / ${formatGas(blockData.gas_limit)} (${((blockData.gas_used / blockData.gas_limit) * 100).toFixed(2)}%)`}
            icon={<Fuel className="w-4 h-4" />}
          />
          <InfoRow
            label="Block Size"
            value={formatBytes(parseInt(blockData.size))}
          />
        </div>
      </div>
    );
  }

  if (type === 'transaction') {
    const txData = data as TransactionResponse;
    return (
      <div className="card">
        <div className="card-header">
          <div className="flex items-center space-x-2">
            <Hash className="w-5 h-5 text-primary-600" />
            <h3 className="text-lg font-semibold text-secondary-900">
              Transaction Details
            </h3>
          </div>
          {txData.status !== undefined && (
            <div className={`flex items-center space-x-1 px-2 py-1 rounded-full text-xs font-medium ${
              txData.status === 1 
                ? 'bg-green-100 text-green-800' 
                : 'bg-red-100 text-red-800'
            }`}>
              {txData.status === 1 ? (
                <CheckCircle className="w-3 h-3" />
              ) : (
                <XCircle className="w-3 h-3" />
              )}
              {txData.status === 1 ? 'Success' : 'Failed'}
            </div>
          )}
        </div>
        
        <div className="space-y-0">
          <InfoRow
            label="Transaction Hash"
            value={txData.hash}
            copyable={txData.hash}
            icon={<Hash className="w-4 h-4" />}
          />
          {txData.block_number && (
            <InfoRow
              label="Block Number"
              value={formatBlockNumber(txData.block_number)}
              icon={<Blocks className="w-4 h-4" />}
            />
          )}
          <InfoRow
            label="From"
            value={txData.from}
            copyable={txData.from}
            icon={<User className="w-4 h-4" />}
          />
          {txData.to && (
            <InfoRow
              label="To"
              value={txData.to}
              copyable={txData.to}
              icon={<User className="w-4 h-4" />}
            />
          )}
          <InfoRow
            label="Value"
            value={`${formatEther(txData.value)} ETH`}
            icon={<DollarSign className="w-4 h-4" />}
          />
          <InfoRow
            label="Gas Limit"
            value={formatGas(txData.gas)}
            icon={<Fuel className="w-4 h-4" />}
          />
          <InfoRow
            label="Gas Price"
            value={`${formatEther(txData.gas_price, 9)} ETH (${(parseFloat(formatEther(txData.gas_price, 9)) * 1e9).toFixed(2)} Gwei)`}
            icon={<Fuel className="w-4 h-4" />}
          />
          {txData.gas_used && (
            <InfoRow
              label="Gas Used"
              value={`${formatGas(txData.gas_used)} (${((txData.gas_used / txData.gas) * 100).toFixed(2)}%)`}
              icon={<Fuel className="w-4 h-4" />}
            />
          )}
        </div>
      </div>
    );
  }

  if (type === 'address') {
    const addressData = data as BalanceResponse & { contractInfo?: ContractInfoResponse };
    const isContract = addressData.contractInfo?.is_contract || false;
    
    return (
      <div className="space-y-6">
        {/* Address Info */}
        <div className="card">
          <div className="card-header">
            <div className="flex items-center space-x-2">
              <User className="w-5 h-5 text-primary-600" />
              <h3 className="text-lg font-semibold text-secondary-900">
                {isContract ? 'Contract' : 'Address'} Details
              </h3>
            </div>
            {isContract && (
              <div className="px-2 py-1 bg-purple-100 text-purple-800 rounded-full text-xs font-medium">
                Smart Contract
              </div>
            )}
          </div>
          
          <div className="space-y-0">
            <InfoRow
              label="Address"
              value={addressData.address}
              copyable={addressData.address}
              icon={<Hash className="w-4 h-4" />}
            />
            <InfoRow
              label="Balance"
              value={`${addressData.balance_eth} ETH`}
              icon={<DollarSign className="w-4 h-4" />}
            />
            <InfoRow
              label="Balance (Wei)"
              value={addressData.balance}
              copyable={addressData.balance}
            />
            {isContract && addressData.contractInfo && (
              <InfoRow
                label="Contract Code Size"
                value={formatBytes(addressData.contractInfo.code_size)}
              />
            )}
          </div>
        </div>

        {/* Contract Code (if applicable) */}
        {isContract && addressData.contractInfo && addressData.contractInfo.code && (
          <div className="card">
            <div className="card-header">
              <h3 className="text-lg font-semibold text-secondary-900">Contract Bytecode</h3>
              <CopyButton text={addressData.contractInfo.code} label="bytecode" />
            </div>
            <div className="bg-secondary-50 rounded-lg p-4 max-h-64 overflow-y-auto">
              <code className="text-xs font-mono text-secondary-700 break-all">
                {addressData.contractInfo.code.length > 1000 
                  ? `${addressData.contractInfo.code.substring(0, 1000)}...` 
                  : addressData.contractInfo.code
                }
              </code>
            </div>
            {addressData.contractInfo.code.length > 1000 && (
              <p className="text-sm text-secondary-500 mt-2">
                Showing first 1000 characters. Click copy to get full bytecode.
              </p>
            )}
          </div>
        )}
      </div>
    );
  }

  return null;
};

export default SearchResults;