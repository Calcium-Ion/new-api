import React, { useEffect, useState } from 'react';

import { getFooterHTML, getSystemName } from '../helpers';
import { Layout, Tooltip } from '@douyinfe/semi-ui';

const FooterBar = () => {
  const systemName = getSystemName();
  const [footer, setFooter] = useState(getFooterHTML());
  let remainCheckTimes = 5;

  const loadFooter = () => {
    let footer_html = localStorage.getItem('footer_html');
    if (footer_html) {
      setFooter(footer_html);
    }
  };

  const defaultFooter = (
    <div className='custom-footer'>
      <a
        href='https://github.com/wmc1125/bigmodel-api'
        target='_blank'
        rel='noreferrer'
      >
        BigModel Api {import.meta.env.VITE_REACT_APP_VERSION}{' '}
      </a>
      由{' '}
      <a
        href='https://github.com/wmc1125'
        target='_blank'
        rel='noreferrer'
      >
        子枫
      </a>{' '}
      开发，基于{' '}
      <a
        href='https://github.com/songquanpeng/one-api'
        target='_blank'
        rel='noreferrer'
      >
        One API
      </a>{' '}
      和{' '}
      <a
        href='https://github.com/Calcium-Ion/new-api'
        target='_blank'
        rel='noreferrer'
      >
        New Api
      </a>
    </div>
  );

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
    <div style={{ textAlign: 'center' }}>
      {footer ? (
        <div
          className='custom-footer'
          dangerouslySetInnerHTML={{ __html: footer }}
        ></div>
      ) : (
        defaultFooter
      )}
    </div>
  );
};

export default FooterBar;
