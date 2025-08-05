import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import Onboard from '@web3-onboard/core';
import injectedModule from '@web3-onboard/injected-wallets';
import walletConnectModule from '@web3-onboard/walletconnect';
import ledgerModule from '@web3-onboard/ledger';
import { ethers } from 'ethers';

const injected = injectedModule();
const walletConnect = walletConnectModule();
const ledger = ledgerModule();

const onboard = Onboard({
  wallets: [injected, walletConnect, ledger],
  chains: [
    {
      id: process.env.REACT_APP_NETWORK_ID || '0x1',
      token: 'ETH',
      label: 'Ethereum',
      rpcUrl: process.env.REACT_APP_RPC_URL || 'https://mainnet.infura.io/v3/YOUR_PROJECT_ID',
    },
    // Add Kairos or other chains here
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
    const previouslyConnectedWallets = onboard.state.get().wallets;
    if (previouslyConnectedWallets.length > 0) {
      onboard.connectWallet();
    }
    const unsubscribe = onboard.state.select('wallets').subscribe(wallets => {
      if (wallets && wallets.length > 0) {
        setAddress(wallets[0].accounts[0].address);
        setProvider(new ethers.BrowserProvider(wallets[0].provider, 'any'));
      } else {
        setAddress(null);
        setProvider(null);
      }
    });
    return () => unsubscribe();
  }, []);

  const connect = async () => {
    setConnecting(true);
    await onboard.connectWallet();
    setConnecting(false);
  };

  const disconnect = async () => {
    const [primary] = onboard.state.get().wallets;
    if (primary) {
      await onboard.disconnectWallet({ label: primary.label });
    }
  };

  const value = useMemo(() => ({ address, provider, onboard, connecting, connect, disconnect }), [address, provider, connecting]);

  return <Web3Context.Provider value={value}>{children}</Web3Context.Provider>;
};

export const useWeb3 = () => {
  const ctx = useContext(Web3Context);
  if (!ctx) throw new Error('useWeb3 must be used within a Web3Provider');
  return ctx;
};