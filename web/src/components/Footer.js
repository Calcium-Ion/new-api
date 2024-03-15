import React, { useEffect, useState } from 'react';

import { getFooterHTML, getSystemName } from '../helpers';
import { Layout } from '@douyinfe/semi-ui';

const Footer = () => {
  const systemName = getSystemName();
  const [footer, setFooter] = useState(getFooterHTML());
  let remainCheckTimes = 5;

  const loadFooter = () => {
    let footer_html = localStorage.getItem('footer_html');
    if (footer_html) {
      setFooter(footer_html);
    }
  };

  useEffect(() => {
    const timer = setInterval(() => {
      if (remainCheckTimes <= 0) {
        clearInterval(timer);
        return;
      }
      remainCheckTimes--;
      loadFooter();
    }, 200);
    return () => clearTimeout(timer);
  }, []);

  return (
    <Layout>
      <Layout.Content style={{ textAlign: 'center' }}>
        {footer ? (
          <div
            className="custom-footer"
            dangerouslySetInnerHTML={{ __html: footer }}
          ></div>
        ) : (
          <div className="custom-footer">
            <a
              href="https://github.com/Calcium-Ion/new-api"
              target="_blank" rel="noreferrer"
            >
              New API {process.env.REACT_APP_VERSION}{' '}
            </a>
            由{' '}
            <a href="https://github.com/Calcium-Ion" target="_blank" rel="noreferrer">
              Calcium-Ion
            </a>{' '}
            开发，基于{' '}
            <a href="https://github.com/songquanpeng/one-api" target="_blank" rel="noreferrer">
              One API v0.5.4
            </a>{' '}
            ，本项目根据{' '}
            <a href="https://opensource.org/licenses/mit-license.php">
              MIT 许可证
            </a>{' '}
            授权
          </div>
        )}
      </Layout.Content>
    </Layout>
  );
};

export default Footer;
