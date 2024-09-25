import React, { useContext, useEffect, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { UserContext } from '../context/User';
import { StatusContext } from '../context/Status';

import {
  API,
  getLogo,
  getSystemName,
  isAdmin,
  isMobile,
  showError,
} from '../helpers';
import '../index.css';

import {
  IconCalendarClock, IconChecklistStroked,
  IconComment, IconCommentStroked,
  IconCreditCard,
  IconGift, IconHelpCircle,
  IconHistogram,
  IconHome,
  IconImage,
  IconKey,
  IconLayers,
  IconPriceTag,
  IconSetting,
  IconUser
} from '@douyinfe/semi-icons';
import { Avatar, Dropdown, Layout, Nav, Switch } from '@douyinfe/semi-ui';
import { setStatusData } from '../helpers/data.js';
import { stringToColor } from '../helpers/render.js';
import { useSetTheme, useTheme } from '../context/Theme/index.js';

// HeaderBar Buttons

const SiderBar = () => {
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const defaultIsCollapsed =
    isMobile() || localStorage.getItem('default_collapse_sidebar') === 'true';

  let navigate = useNavigate();
  const [selectedKeys, setSelectedKeys] = useState(['home']);
  const systemName = getSystemName();
  const logo = getLogo();
  const [isCollapsed, setIsCollapsed] = useState(defaultIsCollapsed);
  const theme = useTheme();
  const setTheme = useSetTheme();

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
    detail: '/detail',
    pricing: '/pricing',
    task: '/task',
    playground: '/playground',
  };

  const headerButtons = useMemo(
    () => [
      // {
      //   text: '首页',
      //   itemKey: 'home',
      //   to: '/',
      //   icon: <IconHome />,
      // },
      {
        text: 'Playground',
        itemKey: 'playground',
        to: '/playground',
        icon: <IconCommentStroked />,
      },
      {
        text: '模型价格',
        itemKey: 'pricing',
        to: '/pricing',
        icon: <IconPriceTag />,
      },
      {
        text: '渠道',
        itemKey: 'channel',
        to: '/channel',
        icon: <IconLayers />,
        className: isAdmin() ? 'semi-navigation-item-normal' : 'tableHiddle',
      },
      {
        text: '聊天',
        itemKey: 'chat',
        to: '/chat',
        icon: <IconComment />,
        className: localStorage.getItem('chat_link')
          ? 'semi-navigation-item-normal'
          : 'tableHiddle',
      },
      {
        text: '令牌',
        itemKey: 'token',
        to: '/token',
        icon: <IconKey />,
      },
      {
        text: '兑换码',
        itemKey: 'redemption',
        to: '/redemption',
        icon: <IconGift />,
        className: isAdmin() ? 'semi-navigation-item-normal' : 'tableHiddle',
      },
      {
        text: '钱包',
        itemKey: 'topup',
        to: '/topup',
        icon: <IconCreditCard />,
      },
      {
        text: '用户管理',
        itemKey: 'user',
        to: '/user',
        icon: <IconUser />,
        className: isAdmin() ? 'semi-navigation-item-normal' : 'tableHiddle',
      },
      {
        text: '日志',
        itemKey: 'log',
        to: '/log',
        icon: <IconHistogram />,
      },
      {
        text: '数据看板',
        itemKey: 'detail',
        to: '/detail',
        icon: <IconCalendarClock />,
        className:
          localStorage.getItem('enable_data_export') === 'true'
            ? 'semi-navigation-item-normal'
            : 'tableHiddle',
      },
      {
        text: '绘图',
        itemKey: 'midjourney',
        to: '/midjourney',
        icon: <IconImage />,
        className:
          localStorage.getItem('enable_drawing') === 'true'
            ? 'semi-navigation-item-normal'
            : 'tableHiddle',
      },
      {
        text: '异步任务',
        itemKey: 'task',
        to: '/task',
        icon: <IconChecklistStroked />,
        className:
            localStorage.getItem('enable_task') === 'true'
                ? 'semi-navigation-item-normal'
                : 'tableHiddle',
      },
      {
        text: '设置',
        itemKey: 'setting',
        to: '/setting',
        icon: <IconSetting />,
      },
      // {
      //     text: '关于',
      //     itemKey: 'about',
      //     to: '/about',
      //     icon: <IconAt/>
      // }
    ],
    [
      localStorage.getItem('enable_data_export'),
      localStorage.getItem('enable_drawing'),
      localStorage.getItem('enable_task'),
      localStorage.getItem('chat_link'),
      isAdmin(),
    ],
  );

  const loadStatus = async () => {
    const res = await API.get('/api/status');
    if (res === undefined) {
      return;
    }
    const { success, data } = res.data;
    if (success) {
      statusDispatch({ type: 'set', payload: data });
      setStatusData(data);
    } else {
      showError('无法正常连接至服务器！');
    }
  };

  useEffect(() => {
    loadStatus().then(() => {
      setIsCollapsed(
        isMobile() ||
          localStorage.getItem('default_collapse_sidebar') === 'true',
      );
    });
    let localKey = window.location.pathname.split('/')[1];
    if (localKey === '') {
      localKey = 'home';
    }
    setSelectedKeys([localKey]);
  }, []);

  return (
    <>
      <Nav
        style={{ maxWidth: 220, height: '100%' }}
        defaultIsCollapsed={
          isMobile() ||
          localStorage.getItem('default_collapse_sidebar') === 'true'
        }
        isCollapsed={isCollapsed}
        onCollapseChange={(collapsed) => {
          setIsCollapsed(collapsed);
        }}
        selectedKeys={selectedKeys}
        renderWrapper={({ itemElement, isSubNav, isInSubNav, props }) => {
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
        onSelect={(key) => {
          setSelectedKeys([key.itemKey]);
        }}
        // header={{
        //   logo: (
        //     <img src={logo} alt='logo' style={{ marginRight: '0.75em' }} />
        //   ),
        //   text: systemName,
        // }}
        // footer={{
        //   text: '© 2021 NekoAPI',
        // }}
        footer={
          <>
            {isMobile() && (
              <Switch
                checkedText='🌞'
                size={'small'}
                checked={theme === 'dark'}
                uncheckedText='🌙'
                onChange={(checked) => {
                  setTheme(checked);
                }}
              />
            )}
          </>
        }
      >
        <Nav.Footer collapseButton={true}></Nav.Footer>
      </Nav>
    </>
  );
};

export default SiderBar;
