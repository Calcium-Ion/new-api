import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import {  Layout } from '@douyinfe/semi-ui';

// 获取 Token Keys 的异步函数，过滤掉非启用状态的令牌
async function fetchTokenKeys() {
  try {
    const response = await API.get('/api/token/?p=0&size=999');
    const { success, data } = response.data;
    if (success) {
      // 过滤已启用状态的令牌
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

const TokenKeysPage = () => {
  const [keys, setKeys] = useState([]);
  const [chatLink, setChatLink] = useState('');
  const [serverAddress, setServerAddress] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  
  useEffect(() => {
    const loadAllData = async () => {
      const fetchedKeys = await fetchTokenKeys();
      if (fetchedKeys.length === 0) {
        // showError('当前没有可用的启用令牌，请确认是否有令牌处于启用状态！');
        setTimeout(() => {
          window.location.href = '/token';
        }, 1500); // 延迟 1.5 秒后跳转
      }
      setKeys(fetchedKeys);
      setIsLoading(false); 

      const link = localStorage.getItem('chat_link');
      setChatLink(link);

      const address = getServerAddress();
      setServerAddress(address);
    };

    loadAllData();
  }, []);

  const comLink = (key) => {
    if (!chatLink || !serverAddress || !key) return '';
    return `${chatLink}/#/?settings={"key":"sk-${key}","url":"${encodeURIComponent(serverAddress)}"}`;
  };

  const iframeSrc = keys.length > 0 ? comLink(keys[0]) : '';

  // 生成链接
  return !isLoading && iframeSrc ? (
    <iframe
      src={iframeSrc}
      style={{ width: '100%', height: '85vh', border: 'none' }}
      title="Token Frame"
    />
  ) : (
    <div>
    <Layout>
    <Layout.Header>
      <h3 style={{ color: 'red'}}>
        当前没有可用的已启用令牌，请确认是否有令牌处于启用状态！<br />
        正在跳转......
      </h3>
    </Layout.Header>
    </Layout>
    </div>
  );

};

export default TokenKeysPage;
