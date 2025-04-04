import React, { useContext, useEffect, useMemo, useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { UserContext } from '../context/User';
import { StatusContext } from '../context/Status';
import { useTranslation } from 'react-i18next';

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
  IconCalendarClock,
  IconChecklistStroked,
  IconComment,
  IconCommentStroked,
  IconCreditCard,
  IconGift,
  IconHelpCircle,
  IconHistogram,
  IconHome,
  IconImage,
  IconKey,
  IconLayers,
  IconPriceTag,
  IconSetting,
  IconUser,
} from '@douyinfe/semi-icons';
import {
  Avatar,
  Dropdown,
  Layout,
  Nav,
  Switch,
  Divider,
} from '@douyinfe/semi-ui';
import { setStatusData } from '../helpers/data.js';
import { stringToColor } from '../helpers/render.js';
import { useSetTheme, useTheme } from '../context/Theme/index.js';
import { StyleContext } from '../context/Style/index.js';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';

// 自定义侧边栏按钮样式
const navItemStyle = {
  borderRadius: '6px',
  margin: '4px 8px',
};

// 自定义侧边栏按钮悬停样式
const navItemHoverStyle = {
  backgroundColor: 'var(--semi-color-primary-light-default)',
  color: 'var(--semi-color-primary)',
};

// 自定义侧边栏按钮选中样式
const navItemSelectedStyle = {
  backgroundColor: 'var(--semi-color-primary-light-default)',
  color: 'var(--semi-color-primary)',
  fontWeight: '600',
};

// 自定义图标样式
const iconStyle = (itemKey, selectedKeys) => {
  return {
    fontSize: '18px',
    color: selectedKeys.includes(itemKey)
      ? 'var(--semi-color-primary)'
      : 'var(--semi-color-text-2)',
  };
};

// Define routerMap as a constant outside the component
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
  detail: '/detail',
  pricing: '/pricing',
  task: '/task',
  playground: '/playground',
  personal: '/personal',
};

const SiderBar = () => {
  const { t } = useTranslation();
  const [styleState, styleDispatch] = useContext(StyleContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const defaultIsCollapsed =
    localStorage.getItem('default_collapse_sidebar') === 'true';

  const [selectedKeys, setSelectedKeys] = useState(['home']);
  const [isCollapsed, setIsCollapsed] = useState(defaultIsCollapsed);
  const [chatItems, setChatItems] = useState([]);
  const [openedKeys, setOpenedKeys] = useState([]);
  const theme = useTheme();
  const setTheme = useSetTheme();
  const location = useLocation();
  const [routerMapState, setRouterMapState] = useState(routerMap);

  // 预先计算所有可能的图标样式
  const allItemKeys = useMemo(() => {
    const keys = [
      'home',
      'channel',
      'token',
      'redemption',
      'topup',
      'user',
      'log',
      'midjourney',
      'setting',
      'about',
      'chat',
      'detail',
      'pricing',
      'task',
      'playground',
      'personal',
    ];
    // 添加聊天项的keys
    for (let i = 0; i < chatItems.length; i++) {
      keys.push('chat' + i);
    }
    return keys;
  }, [chatItems]);

  // 使用useMemo一次性计算所有图标样式
  const iconStyles = useMemo(() => {
    const styles = {};
    allItemKeys.forEach((key) => {
      styles[key] = iconStyle(key, selectedKeys);
    });
    return styles;
  }, [allItemKeys, selectedKeys]);

  const workspaceItems = useMemo(
    () => [
      {
        text: t('数据看板'),
        itemKey: 'detail',
        to: '/detail',
        icon: <IconCalendarClock />,
        className:
          localStorage.getItem('enable_data_export') === 'true'
            ? ''
            : 'tableHiddle',
      },
      {
        text: t('API令牌'),
        itemKey: 'token',
        to: '/token',
        icon: <IconKey />,
      },
      {
        text: t('使用日志'),
        itemKey: 'log',
        to: '/log',
        icon: <IconHistogram />,
      },
      {
        text: t('绘图日志'),
        itemKey: 'midjourney',
        to: '/midjourney',
        icon: <IconImage />,
        className:
          localStorage.getItem('enable_drawing') === 'true'
            ? ''
            : 'tableHiddle',
      },
      {
        text: t('任务日志'),
        itemKey: 'task',
        to: '/task',
        icon: <IconChecklistStroked />,
        className:
          localStorage.getItem('enable_task') === 'true' ? '' : 'tableHiddle',
      },
    ],
    [
      localStorage.getItem('enable_data_export'),
      localStorage.getItem('enable_drawing'),
      localStorage.getItem('enable_task'),
      t,
    ],
  );

  const financeItems = useMemo(
    () => [
      {
        text: t('钱包'),
        itemKey: 'topup',
        to: '/topup',
        icon: <IconCreditCard />,
      },
      {
        text: t('个人设置'),
        itemKey: 'personal',
        to: '/personal',
        icon: <IconUser />,
      },
    ],
    [t],
  );

  const adminItems = useMemo(
    () => [
      {
        text: t('渠道'),
        itemKey: 'channel',
        to: '/channel',
        icon: <IconLayers />,
        className: isAdmin() ? '' : 'tableHiddle',
      },
      {
        text: t('兑换码'),
        itemKey: 'redemption',
        to: '/redemption',
        icon: <IconGift />,
        className: isAdmin() ? '' : 'tableHiddle',
      },
      {
        text: t('用户管理'),
        itemKey: 'user',
        to: '/user',
        icon: <IconUser />,
      },
      {
        text: t('系统设置'),
        itemKey: 'setting',
        to: '/setting',
        icon: <IconSetting />,
      },
    ],
    [isAdmin(), t],
  );

  const chatMenuItems = useMemo(
    () => [
      {
        text: 'Playground',
        itemKey: 'playground',
        to: '/playground',
        icon: <IconCommentStroked />,
      },
      {
        text: t('聊天'),
        itemKey: 'chat',
        items: chatItems,
        icon: <IconComment />,
      },
    ],
    [chatItems, t],
  );

  // Function to update router map with chat routes
  const updateRouterMapWithChats = (chats) => {
    const newRouterMap = { ...routerMap };

    if (Array.isArray(chats) && chats.length > 0) {
      for (let i = 0; i < chats.length; i++) {
        newRouterMap['chat' + i] = '/chat/' + i;
      }
    }

    setRouterMapState(newRouterMap);
    return newRouterMap;
  };

  // Update the useEffect for chat items
  useEffect(() => {
    let chats = localStorage.getItem('chats');
    if (chats) {
      try {
        chats = JSON.parse(chats);
        if (Array.isArray(chats)) {
          let chatItems = [];
          for (let i = 0; i < chats.length; i++) {
            let chat = {};
            for (let key in chats[i]) {
              chat.text = key;
              chat.itemKey = 'chat' + i;
              chat.to = '/chat/' + i;
            }
            chatItems.push(chat);
          }
          setChatItems(chatItems);

          // Update router map with chat routes
          updateRouterMapWithChats(chats);
        }
      } catch (e) {
        console.error(e);
        showError('聊天数据解析失败');
      }
    }
  }, []);

  // Update the useEffect for route selection
  useEffect(() => {
    const currentPath = location.pathname;
    let matchingKey = Object.keys(routerMapState).find(
      (key) => routerMapState[key] === currentPath,
    );

    // Handle chat routes
    if (!matchingKey && currentPath.startsWith('/chat/')) {
      const chatIndex = currentPath.split('/').pop();
      if (!isNaN(chatIndex)) {
        matchingKey = 'chat' + chatIndex;
      } else {
        matchingKey = 'chat';
      }
    }

    // If we found a matching key, update the selected keys
    if (matchingKey) {
      setSelectedKeys([matchingKey]);
    }
  }, [location.pathname, routerMapState]);

  useEffect(() => {
    setIsCollapsed(styleState.siderCollapsed);
  }, [styleState.siderCollapsed]);

  // Custom divider style
  const dividerStyle = {
    margin: '8px 0',
    opacity: 0.6,
  };

  // Custom group label style
  const groupLabelStyle = {
    padding: '8px 16px',
    color: 'var(--semi-color-text-2)',
    fontSize: '12px',
    fontWeight: 'bold',
    textTransform: 'uppercase',
    letterSpacing: '0.5px',
  };

  return (
    <>
      <Nav
        className='custom-sidebar-nav'
        style={{
          width: isCollapsed ? '60px' : '200px',
          boxShadow: '0 2px 8px rgba(0, 0, 0, 0.15)',
          borderRight: '1px solid var(--semi-color-border)',
          background: 'var(--semi-color-bg-1)',
          borderRadius: styleState.isMobile ? '0' : '0 8px 8px 0',
          position: 'relative',
          zIndex: 95,
          height: '100%',
          overflowY: 'auto',
          WebkitOverflowScrolling: 'touch', // Improve scrolling on iOS devices
        }}
        defaultIsCollapsed={
          localStorage.getItem('default_collapse_sidebar') === 'true'
        }
        isCollapsed={isCollapsed}
        onCollapseChange={(collapsed) => {
          setIsCollapsed(collapsed);
          // styleDispatch({ type: 'SET_SIDER', payload: true });
          styleDispatch({ type: 'SET_SIDER_COLLAPSED', payload: collapsed });
          localStorage.setItem('default_collapse_sidebar', collapsed);

          // 确保在收起侧边栏时有选中的项目，避免不必要的计算
          if (selectedKeys.length === 0) {
            const currentPath = location.pathname;
            const matchingKey = Object.keys(routerMapState).find(
              (key) => routerMapState[key] === currentPath,
            );

            if (matchingKey) {
              setSelectedKeys([matchingKey]);
            } else if (currentPath.startsWith('/chat/')) {
              setSelectedKeys(['chat']);
            } else {
              setSelectedKeys(['detail']); // 默认选中首页
            }
          }
        }}
        selectedKeys={selectedKeys}
        itemStyle={navItemStyle}
        hoverStyle={navItemHoverStyle}
        selectedStyle={navItemSelectedStyle}
        renderWrapper={({ itemElement, isSubNav, isInSubNav, props }) => {
          return (
            <Link
              style={{ textDecoration: 'none' }}
              to={routerMapState[props.itemKey] || routerMap[props.itemKey]}
            >
              {itemElement}
            </Link>
          );
        }}
        onSelect={(key) => {
          if (key.itemKey.toString().startsWith('chat')) {
            styleDispatch({ type: 'SET_INNER_PADDING', payload: false });
          } else {
            styleDispatch({ type: 'SET_INNER_PADDING', payload: true });
          }

          // 如果点击的是已经展开的子菜单的父项，则收起子菜单
          if (openedKeys.includes(key.itemKey)) {
            setOpenedKeys(openedKeys.filter((k) => k !== key.itemKey));
          }

          setSelectedKeys([key.itemKey]);
        }}
        openKeys={openedKeys}
        onOpenChange={(data) => {
          setOpenedKeys(data.openKeys);
        }}
      >
        {/* Chat Section - Only show if there are chat items */}
        {chatMenuItems.map((item) => {
          if (item.items && item.items.length > 0) {
            return (
              <Nav.Sub
                key={item.itemKey}
                itemKey={item.itemKey}
                text={item.text}
                icon={React.cloneElement(item.icon, {
                  style: iconStyles[item.itemKey],
                })}
              >
                {item.items.map((subItem) => (
                  <Nav.Item
                    key={subItem.itemKey}
                    itemKey={subItem.itemKey}
                    text={subItem.text}
                  />
                ))}
              </Nav.Sub>
            );
          } else {
            return (
              <Nav.Item
                key={item.itemKey}
                itemKey={item.itemKey}
                text={item.text}
                icon={React.cloneElement(item.icon, {
                  style: iconStyles[item.itemKey],
                })}
              />
            );
          }
        })}

        {/* Divider */}
        <Divider style={dividerStyle} />

        {/* Workspace Section */}
        {!isCollapsed && <Text style={groupLabelStyle}>{t('控制台')}</Text>}
        {workspaceItems.map((item) => (
          <Nav.Item
            key={item.itemKey}
            itemKey={item.itemKey}
            text={item.text}
            icon={React.cloneElement(item.icon, {
              style: iconStyles[item.itemKey],
            })}
            className={item.className}
          />
        ))}

        {isAdmin() && (
          <>
            {/* Divider */}
            <Divider style={dividerStyle} />

            {/* Admin Section */}
            {!isCollapsed && <Text style={groupLabelStyle}>{t('管理员')}</Text>}
            {adminItems.map((item) => (
              <Nav.Item
                key={item.itemKey}
                itemKey={item.itemKey}
                text={item.text}
                icon={React.cloneElement(item.icon, {
                  style: iconStyles[item.itemKey],
                })}
                className={item.className}
              />
            ))}
          </>
        )}

        {/* Divider */}
        <Divider style={dividerStyle} />

        {/* Finance Management Section */}
        {!isCollapsed && <Text style={groupLabelStyle}>{t('个人中心')}</Text>}
        {financeItems.map((item) => (
          <Nav.Item
            key={item.itemKey}
            itemKey={item.itemKey}
            text={item.text}
            icon={React.cloneElement(item.icon, {
              style: iconStyles[item.itemKey],
            })}
            className={item.className}
          />
        ))}

        <Nav.Footer
          style={{
            paddingBottom: styleState?.isMobile ? '112px' : '',
          }}
          collapseButton={true}
          collapseText={(collapsed) => {
            if (collapsed) {
              return t('展开侧边栏');
            }
            return t('收起侧边栏');
          }}
        />
      </Nav>
    </>
  );
};

export default SiderBar;
