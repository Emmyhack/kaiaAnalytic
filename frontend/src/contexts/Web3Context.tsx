import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import Onboard from '@web3-onboard/core';
import injectedModule from '@web3-onboard/injected-wallets';
import walletConnectModule from '@web3-onboard/walletconnect';
import ledgerModule from '@web3-onboard/ledger';
import { ethers } from 'ethers';

const injected = injectedModule();
const walletConnect = walletConnectModule();
const ledger = ledgerModule();

// Environment variables with proper fallbacks
const NETWORK_ID = import.meta.env.VITE_REACT_APP_NETWORK_ID || '8217';
const RPC_URL = import.meta.env.VITE_REACT_APP_RPC_URL || 'https://kaia-mainnet.blockpi.network/v1/rpc/public';

const onboard = Onboard({
  wallets: [injected, walletConnect, ledger],
  chains: [
    {
      id: NETWORK_ID,
      token: 'KAIA',
      label: 'Kaia Mainnet',
      rpcUrl: RPC_URL,
    },
  ],
  appMetadata: {
    name: 'KaiaAnalyticsAI',
    icon: '<svg></svg>',
    description: 'Decentralized Analytics Platform',
    recommendedInjectedWallets: [
      { name: 'MetaMask', url: 'https://metamask.io' },
      { name: 'Coinbase', url: 'https://wallet.coinbase.com/' }
    ]
  }
});

interface Web3ContextProps {
  address: string | null;
  provider: ethers.BrowserProvider | null;
  onboard: typeof onboard;
  connecting: boolean;
  connect: () => Promise<void>;
  disconnect: () => Promise<void>;
}

const Web3Context = createContext<Web3ContextProps | undefined>(undefined);

export const Web3Provider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [address, setAddress] = useState<string | null>(null);
  const [provider, setProvider] = useState<ethers.BrowserProvider | null>(null);
  const [connecting, setConnecting] = useState(false);

  useEffect(() => {
    // Subscribe to wallet updates
    const wallets = onboard.state.select('wallets');
    
    const { unsubscribe } = wallets.subscribe(wallets => {
      if (wallets[0]) {
        setAddress(wallets[0].accounts[0]?.address || null);
        
        if (wallets[0].provider) {
          setProvider(new ethers.BrowserProvider(wallets[0].provider));
        }
      } else {
        setAddress(null);
        setProvider(null);
      }
    });

    return () => unsubscribe();
  }, []);

  const connect = async () => {
    try {
      setConnecting(true);
      await onboard.connectWallet();
    } catch (error) {
      console.error('Failed to connect wallet:', error);
    } finally {
      setConnecting(false);
    }
  };

  const disconnect = async () => {
    const [primaryWallet] = onboard.state.get().wallets;
    if (primaryWallet) {
      await onboard.disconnectWallet({ label: primaryWallet.label });
    }
  };

  const value = useMemo(
    () => ({
      address,
      provider,
      onboard,
      connecting,
      connect,
      disconnect,
    }),
    [address, provider, connecting]
  );

  return (
    <Web3Context.Provider value={value}>
      {children}
    </Web3Context.Provider>
  );
};

export const useWeb3 = () => {
  const context = useContext(Web3Context);
  if (context === undefined) {
    throw new Error('useWeb3 must be used within a Web3Provider');
  }
  return context;
};