import React, { useState, useEffect, useRef } from 'react';
import { PaperAirplaneIcon, SparklesIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';
import { api } from '../services/api';

function Chat() {
  const [messages, setMessages] = useState([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [suggestedQueries] = useState([
    'What are the best yield opportunities?',
    'Show me trading suggestions for KAIA',
    'What governance proposals are active?',
    'How do I stake my KAIA tokens?',
    'What is the current gas price trend?',
  ]);

  const messagesEndRef = useRef(null);
  const wsRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    // Initialize WebSocket connection
    const ws = new WebSocket('ws://localhost:8080/api/v1/chat/ws');
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      toast.success('Connected to chat');
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'response') {
        setMessages(prev => [...prev, {
          id: Date.now(),
          type: 'bot',
          content: data.content,
          data: data.data,
          timestamp: new Date(),
        }]);
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      toast.error('Disconnected from chat');
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      toast.error('Connection error');
    };

    return () => {
      ws.close();
    };
  }, []);

  const sendMessage = async (message) => {
    if (!message.trim()) return;

    const userMessage = {
      id: Date.now(),
      type: 'user',
      content: message,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputMessage('');
    setIsLoading(true);

    try {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({
          type: 'query',
          content: message,
          timestamp: Date.now(),
        }));
      } else {
        // Fallback to HTTP API
        const response = await api.post('/api/v1/chat/query', {
          query: message,
        });

        setMessages(prev => [...prev, {
          id: Date.now(),
          type: 'bot',
          content: response.data.answer,
          data: response.data,
          timestamp: new Date(),
        }]);
      }
    } catch (error) {
      console.error('Error sending message:', error);
      toast.error('Failed to send message');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    sendMessage(inputMessage);
  };

  const executeAction = async (action) => {
    try {
      toast.loading('Executing action...');
      
      // In a real implementation, this would trigger a blockchain transaction
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      toast.success(`Action executed: ${action.description}`);
      
      setMessages(prev => [...prev, {
        id: Date.now(),
        type: 'system',
        content: `✅ Action executed: ${action.description}`,
        timestamp: new Date(),
      }]);
    } catch (error) {
      console.error('Error executing action:', error);
      toast.error('Failed to execute action');
    }
  };

  const formatTimestamp = (timestamp) => {
    return timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="flex flex-col h-[calc(100vh-200px)]">
      {/* Header */}
      <div className="bg-white shadow rounded-lg mb-6">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">AI Chat Assistant</h1>
              <p className="mt-1 text-sm text-gray-500">
                Ask me anything about Kaia analytics, trading, or blockchain actions
              </p>
            </div>
            <div className="flex items-center space-x-2">
              <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-400' : 'bg-red-400'}`} />
              <span className="text-sm text-gray-500">
                {isConnected ? 'Connected' : 'Disconnected'}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 flex">
        {/* Chat area */}
        <div className="flex-1 flex flex-col">
          <div className="bg-white shadow rounded-lg flex-1 flex flex-col">
            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4">
              {messages.length === 0 && (
                <div className="text-center text-gray-500">
                  <SparklesIcon className="mx-auto h-12 w-12 text-gray-400" />
                  <h3 className="mt-2 text-sm font-medium text-gray-900">Start a conversation</h3>
                  <p className="mt-1 text-sm text-gray-500">
                    Ask me about yield opportunities, trading suggestions, or blockchain actions.
                  </p>
                </div>
              )}

              {messages.map((message) => (
                <div
                  key={message.id}
                  className={`flex ${message.type === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                      message.type === 'user'
                        ? 'bg-blue-600 text-white'
                        : message.type === 'system'
                        ? 'bg-green-100 text-green-800'
                        : 'bg-gray-100 text-gray-900'
                    }`}
                  >
                    <p className="text-sm">{message.content}</p>
                    <p className="text-xs opacity-75 mt-1">
                      {formatTimestamp(message.timestamp)}
                    </p>
                  </div>
                </div>
              ))}

              {/* Action buttons for bot messages */}
              {messages.map((message) => 
                message.type === 'bot' && message.data?.actions ? (
                  <div key={`actions-${message.id}`} className="flex justify-start">
                    <div className="max-w-xs lg:max-w-md">
                      <div className="text-xs text-gray-500 mb-2">Suggested actions:</div>
                      <div className="space-y-2">
                        {message.data.actions.map((action, index) => (
                          <button
                            key={index}
                            onClick={() => executeAction(action)}
                            className="w-full text-left px-3 py-2 text-sm bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-md transition-colors"
                          >
                            <div className="font-medium">{action.description}</div>
                            <div className="text-xs opacity-75">
                              Confidence: {(action.confidence * 100).toFixed(0)}%
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>
                ) : null
              )}

              {isLoading && (
                <div className="flex justify-start">
                  <div className="max-w-xs lg:max-w-md px-4 py-2 rounded-lg bg-gray-100">
                    <div className="flex items-center space-x-2">
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600"></div>
                      <span className="text-sm text-gray-600">AI is thinking...</span>
                    </div>
                  </div>
                </div>
              )}

              <div ref={messagesEndRef} />
            </div>

            {/* Input area */}
            <div className="border-t border-gray-200 p-4">
              <form onSubmit={handleSubmit} className="flex space-x-4">
                <input
                  type="text"
                  value={inputMessage}
                  onChange={(e) => setInputMessage(e.target.value)}
                  placeholder="Ask me anything..."
                  className="flex-1 rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                  disabled={!isConnected}
                />
                <button
                  type="submit"
                  disabled={!isConnected || isLoading || !inputMessage.trim()}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <PaperAirplaneIcon className="h-4 w-4" />
                </button>
              </form>
            </div>
          </div>
        </div>

        {/* Suggested queries */}
        <div className="hidden lg:block w-80 ml-6">
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Suggested Queries</h3>
            <div className="space-y-3">
              {suggestedQueries.map((query, index) => (
                <button
                  key={index}
                  onClick={() => sendMessage(query)}
                  className="w-full text-left p-3 text-sm bg-gray-50 hover:bg-gray-100 text-gray-700 rounded-md transition-colors"
                >
                  {query}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Chat;