// src/hooks/useTokenKeys.js
import { useEffect, useState } from 'react';
import { API, showError } from '../helpers';

async function fetchTokenKeys() {
  try {
    const response = await API.get('/api/token/?p=0&size=100');
    const { success, data } = response.data;
    if (success) {
      const activeTokens = data.filter((token) => token.status === 1);
      return activeTokens.map((token) => token.key);
    } else {
      throw new Error('Failed to fetch token keys');
    }
  } catch (error) {
    console.error("Error fetching token keys:", error);
    return [];
  }
}

function getServerAddress() {
  let status = localStorage.getItem('status');
  let serverAddress = '';

  if (status) {
    try {
      status = JSON.parse(status);
      serverAddress = status.server_address || '';
    } catch (error) {
      console.error("Failed to parse status from localStorage:", error);
    }
  }

  if (!serverAddress) {
    serverAddress = window.location.origin;
  }

  return serverAddress;
}

export function useTokenKeys(id) {
  const [keys, setKeys] = useState([]);
  // const [chatLink, setChatLink] = useState('');
  const [serverAddress, setServerAddress] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadAllData = async () => {
      const fetchedKeys = await fetchTokenKeys();
      if (fetchedKeys.length === 0) {
        showError('当前没有可用的启用令牌，请确认是否有令牌处于启用状态！');
        setTimeout(() => {
          window.location.href = '/token';
        }, 1500); // 延迟 1.5 秒后跳转
      }
      setKeys(fetchedKeys);
      setIsLoading(false);
      // setChatLink(link);

      const address = getServerAddress();
      setServerAddress(address);
    };

    loadAllData();
  }, []);

  return { keys, serverAddress, isLoading };
}