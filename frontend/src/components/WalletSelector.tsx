import React from 'react';
import { useWeb3 } from '../contexts/Web3Context';

const WalletSelector: React.FC = () => {
  const { address, connect, disconnect, connecting, onboard } = useWeb3();

  return (
    <div className="flex flex-col items-center space-y-2">
      {address ? (
        <>
          <div className="text-green-600 font-mono text-sm">Connected: {address.slice(0, 6)}...{address.slice(-4)}</div>
          <button
            className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
            onClick={disconnect}
          >
            Disconnect
          </button>
        </>
      ) : (
        <button
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          onClick={connect}
          disabled={connecting}
        >
          {connecting ? 'Connecting...' : 'Connect Wallet'}
        </button>
      )}
      {/* Advanced selector: show all available wallets */}
      {!address && (
        <div className="mt-2 grid grid-cols-2 gap-2">
          {onboard.state.get().wallets.length === 0 && onboard.state.get().walletModules.map((mod, i) => (
            <button
              key={mod.label}
              className="px-2 py-1 border rounded text-xs hover:bg-gray-100"
              onClick={() => onboard.connectWallet({ autoSelect: { label: mod.label, disableModals: true } })}
            >
              {mod.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default WalletSelector;