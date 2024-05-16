import React, { useEffect, useState } from 'react';
import { Table, Tag, Space } from 'antd';
import { API } from '../../helpers';

const ChannelErrors = () => {
  const [loading, setLoading] = useState(false);
  const [channelErrors, setChannelErrors] = useState([]);

  useEffect(() => {
    const fetchChannelErrors = async () => {
      setLoading(true);
      try {
        const response = await API.get('/admin/channel-errors');
        setChannelErrors(response.data);
      } catch (error) {
        console.error('Failed to fetch channel errors:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchChannelErrors();
  }, []);

  const columns = [
    {
      title: 'Channel ID',
      dataIndex: 'channelId',
      key: 'channelId',
    },
    {
      title: 'Error Message',
      dataIndex: 'errorMessage',
      key: 'errorMessage',
    },
    {
      title: 'Timestamp',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: timestamp => new Date(timestamp).toLocaleString(),
    },
  ];

  return (
    <div>
      <h2>Channel Errors</h2>
      <Table columns={columns} dataSource={channelErrors} loading={loading} />
    </div>
  );
};

export default ChannelErrors;
