import React from 'react';
import { Spin } from '@douyinfe/semi-ui';

const Loading = ({ prompt: name = 'page' }) => {
  return (
    <Spin style={{ height: 100 }} spinning={true}>
      加载{name}中...
    </Spin>
  );
};

export default Loading;
