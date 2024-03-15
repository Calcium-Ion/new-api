import React, { useContext, useEffect, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { UserContext } from '../context/User';
import { StatusContext } from '../context/Status';

import { API, getLogo, getSystemName, isAdmin, isMobile, showError } from '../helpers';
import '../index.css';

import {
  IconCalendarClock,
  IconComment,
  IconCreditCard,
  IconGift,
  IconHistogram,
  IconHome,
  IconImage,
  IconKey,
  IconLayers,
  IconSetting,
  IconUser
} from '@douyinfe/semi-icons';
import { Layout, Nav } from '@douyinfe/semi-ui';

// HeaderBar Buttons

const SiderBar = () => {
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const defaultIsCollapsed = isMobile() || localStorage.getItem('default_collapse_sidebar') === 'true';

  let navigate = useNavigate();
  const [selectedKeys, setSelectedKeys] = useState(['home']);
  const systemName = getSystemName();
  const logo = getLogo();
  const [isCollapsed, setIsCollapsed] = useState(defaultIsCollapsed);

  const headerButtons = useMemo(() => [
    {
      text: '首页',
      itemKey: 'home',
      to: '/',
      icon: <IconHome />
    },
    {
      text: '渠道',
      itemKey: 'channel',
      to: '/channel',
      icon: <IconLayers />,
      className: isAdmin() ? 'semi-navigation-item-normal' : 'tableHiddle'
    },
    {
      text: '聊天',
      itemKey: 'chat',
      to: '/chat',
      icon: <IconComment />,
      className: localStorage.getItem('chat_link') ? 'semi-navigation-item-normal' : 'tableHiddle'
    },
    {
      text: '令牌',
      itemKey: 'token',
      to: '/token',
      icon: <IconKey />
    },
    {
      text: '兑换码',
      itemKey: 'redemption',
      to: '/redemption',
      icon: <IconGift />,
      className: isAdmin() ? 'semi-navigation-item-normal' : 'tableHiddle'
    },
    {
      text: '钱包',
      itemKey: 'topup',
      to: '/topup',
      icon: <IconCreditCard />
    },
    {
      text: '用户管理',
      itemKey: 'user',
      to: '/user',
      icon: <IconUser />,
      className: isAdmin() ? 'semi-navigation-item-normal' : 'tableHiddle'
    },
    {
      text: '日志',
      itemKey: 'log',
      to: '/log',
      icon: <IconHistogram />
    },
    {
      text: '数据看板',
      itemKey: 'detail',
      to: '/detail',
      icon: <IconCalendarClock />,
      className: localStorage.getItem('enable_data_export') === 'true' ? 'semi-navigation-item-normal' : 'tableHiddle'
    },
    {
      text: '绘图',
      itemKey: 'midjourney',
      to: '/midjourney',
      icon: <IconImage />,
      className: localStorage.getItem('enable_drawing') === 'true' ? 'semi-navigation-item-normal' : 'tableHiddle'
    },
    {
      text: '设置',
      itemKey: 'setting',
      to: '/setting',
      icon: <IconSetting />
    }
    // {
    //     text: '关于',
    //     itemKey: 'about',
    //     to: '/about',
    //     icon: <IconAt/>
    // }
  ], [localStorage.getItem('enable_data_export'), localStorage.getItem('enable_drawing'), localStorage.getItem('chat_link'), isAdmin()]);

  const loadStatus = async () => {
    const res = await API.get('/api/status');
    const { success, data } = res.data;
    if (success) {
      localStorage.setItem('status', JSON.stringify(data));
      statusDispatch({ type: 'set', payload: data });
      localStorage.setItem('system_name', data.system_name);
      localStorage.setItem('logo', data.logo);
      localStorage.setItem('footer_html', data.footer_html);
      localStorage.setItem('quota_per_unit', data.quota_per_unit);
      localStorage.setItem('display_in_currency', data.display_in_currency);
      localStorage.setItem('enable_drawing', data.enable_drawing);
      localStorage.setItem('enable_data_export', data.enable_data_export);
      localStorage.setItem('data_export_default_time', data.data_export_default_time);
      localStorage.setItem('default_collapse_sidebar', data.default_collapse_sidebar);
      localStorage.setItem('mj_notify_enabled', data.mj_notify_enabled);
      if (data.chat_link) {
        localStorage.setItem('chat_link', data.chat_link);
      } else {
        localStorage.removeItem('chat_link');
      }
      if (data.chat_link2) {
        localStorage.setItem('chat_link2', data.chat_link2);
      } else {
        localStorage.removeItem('chat_link2');
      }
    } else {
      showError('无法正常连接至服务器！');
    }
  };

  useEffect(() => {
    loadStatus().then(() => {
      setIsCollapsed(isMobile() || localStorage.getItem('default_collapse_sidebar') === 'true');
    });
  }, []);

  return (
    <>
      <Layout>
        <div style={{ height: '100%' }}>
          <Nav
            // bodyStyle={{ maxWidth: 200 }}
            style={{ maxWidth: 200 }}
            defaultIsCollapsed={isMobile() || localStorage.getItem('default_collapse_sidebar') === 'true'}
            isCollapsed={isCollapsed}
            onCollapseChange={collapsed => {
              setIsCollapsed(collapsed);
            }}
            selectedKeys={selectedKeys}
            renderWrapper={({ itemElement, isSubNav, isInSubNav, props }) => {
              const routerMap = {
                home: '/',
                channel: '/channel',
                token: '/token',
                redemption: '/redemption',
                topup: '/topup',
                user: '/user',
                log: '/log',
                midjourney: '/midjourney',
                setting: '/setting',
                about: '/about',
                chat: '/chat',
                detail: '/detail'
              };
              return (
                <Link
                  style={{ textDecoration: 'none' }}
                  to={routerMap[props.itemKey]}
                >
                  {itemElement}
                </Link>
              );
            }}
            items={headerButtons}
            onSelect={key => {
              setSelectedKeys([key.itemKey]);
            }}
            header={{
              logo: <img src={logo} alt="logo" style={{ marginRight: '0.75em' }} />,
              text: systemName
            }}
            // footer={{
            //   text: '© 2021 NekoAPI',
            // }}
          >

            <Nav.Footer collapseButton={true}>
            </Nav.Footer>
          </Nav>
        </div>
      </Layout>
    </>
  );
};

export default SiderBar;
