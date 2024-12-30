import React, { useCallback, useContext, useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { UserContext } from '../../context/User/index.js';
import { API, getUserIdFromLocalStorage, showError } from '../../helpers/index.js';
import { Card, Chat, Input, Layout, Select, Slider, TextArea, Typography, Button, Highlight } from '@douyinfe/semi-ui';
import { SSE } from 'sse';
import { IconSetting } from '@douyinfe/semi-icons';
import { StyleContext } from '../../context/Style/index.js';
import { useTranslation } from 'react-i18next';
import { renderGroupOption } from '../../helpers/render.js';

const roleInfo = {
  user: {
    name: 'User',
    avatar: 'https://lf3-static.bytednsdoc.com/obj/eden-cn/ptlz_zlp/ljhwZthlaukjlkulzlp/docs-icon.png'
  },
  assistant: {
    name: 'Assistant',
    avatar: 'logo.png'
  },
  system: {
    name: 'System',
    avatar: 'https://lf3-static.bytednsdoc.com/obj/eden-cn/ptlz_zlp/ljhwZthlaukjlkulzlp/other/logo.png'
  }
}

let id = 4;
function getId() {
  return `${id++}`
}

const Playground = () => {
  const { t } = useTranslation();
  
  const defaultMessage = [
    {
      role: 'user',
      id: '2',
      createAt: 1715676751919,
      content: t('你好'),
    },
    {
      role: 'assistant',
      id: '3',
      createAt: 1715676751919,
      content: t('你好，请问有什么可以帮助您的吗？'),
    }
  ];

  const [inputs, setInputs] = useState({
    model: 'gpt-4o-mini',
    group: '',
    max_tokens: 0,
    temperature: 0,
  });
  const [searchParams, setSearchParams] = useSearchParams();
  const [userState, userDispatch] = useContext(UserContext);
  const [status, setStatus] = useState({});
  const [systemPrompt, setSystemPrompt] = useState('You are a helpful assistant. You can help me by answering my questions. You can also ask me questions.');
  const [message, setMessage] = useState(defaultMessage);
  const [models, setModels] = useState([]);
  const [groups, setGroups] = useState([]);
  const [showSettings, setShowSettings] = useState(true);
  const [styleState, styleDispatch] = useContext(StyleContext);

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  useEffect(() => {
    if (searchParams.get('expired')) {
      showError(t('未登录或登录已过期，请重新登录！'));
    }
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      setStatus(status);
    }
    loadModels();
    loadGroups();
  }, []);

  const loadModels = async () => {
    let res = await API.get(`/api/user/models`);
    const { success, message, data } = res.data;
    if (success) {
      let localModelOptions = data.map((model) => ({
        label: model,
        value: model,
      }));
      setModels(localModelOptions);
    } else {
      showError(t(message));
    }
  };

  const loadGroups = async () => {
    let res = await API.get(`/api/user/self/groups`);
    const { success, message, data } = res.data;
    if (success) {
      let localGroupOptions = Object.entries(data).map(([group, info]) => ({
        label: info.desc,
        value: group,
        ratio: info.ratio
      }));

      if (localGroupOptions.length === 0) {
        localGroupOptions = [{
          label: t('用户分组'),
          value: '',
          ratio: 1
        }];
      } else {
        const localUser = JSON.parse(localStorage.getItem('user'));
        const userGroup = (userState.user && userState.user.group) || (localUser && localUser.group);
        
        if (userGroup) {
          const userGroupIndex = localGroupOptions.findIndex(g => g.value === userGroup);
          if (userGroupIndex > -1) {
            const userGroupOption = localGroupOptions.splice(userGroupIndex, 1)[0];
            localGroupOptions.unshift(userGroupOption);
          }
        }
      }

      setGroups(localGroupOptions);
      handleInputChange('group', localGroupOptions[0].value);
    } else {
      showError(t(message));
    }
  };

  const commonOuterStyle = {
    border: '1px solid var(--semi-color-border)',
    borderRadius: '16px',
    margin: '0px 8px',
  }

  const getSystemMessage = () => {
    if (systemPrompt !== '') {
      return {
        role: 'system',
        id: '1',
        createAt: 1715676751919,
        content: systemPrompt,
      }
    }
  }

  let handleSSE = (payload) => {
    let source = new SSE('/pg/chat/completions', {
      headers: {
        "Content-Type": "application/json",
        "New-Api-User": getUserIdFromLocalStorage(),
      },
      method: "POST",
      payload: JSON.stringify(payload),
    });
    source.addEventListener("message", (e) => {
      if (e.data !== "[DONE]") {
        let payload = JSON.parse(e.data);
        // console.log("Payload: ", payload);
        if (payload.choices.length === 0) {
          source.close();
          completeMessage();
        } else {
          let text = payload.choices[0].delta.content;
          if (text) {
            generateMockResponse(text);
          }
        }
      } else {
        completeMessage();
      }
    });

    source.addEventListener("error", (e) => {
      generateMockResponse(e.data)
      completeMessage('error')
    });

    source.addEventListener("readystatechange", (e) => {
      if (e.readyState >= 2) {
        if (source.status === undefined) {
          source.close();
          completeMessage();
        }
      }
    });
    source.stream();
  }

  const onMessageSend = useCallback((content, attachment) => {
    console.log("attachment: ", attachment);
    setMessage((prevMessage) => {
      const newMessage = [
        ...prevMessage,
        {
          role: 'user',
          content: content,
          createAt: Date.now(),
          id: getId()
        }
      ];

      // 将 getPayload 移到这里
      const getPayload = () => {
        let systemMessage = getSystemMessage();
        let messages = newMessage.map((item) => {
          return {
            role: item.role,
            content: item.content,
          }
        });
        if (systemMessage) {
          messages.unshift(systemMessage);
        }
        return {
          messages: messages,
          stream: true,
          model: inputs.model,
          group: inputs.group,
          max_tokens: parseInt(inputs.max_tokens),
          temperature: inputs.temperature,
        };
      };

      // 使用更新后的消息状态调用 handleSSE
      handleSSE(getPayload());
      newMessage.push({
        role: 'assistant',
        content: '',
        createAt: Date.now(),
        id: getId(),
        status: 'loading'
      });
      return newMessage;
    });
  }, [getSystemMessage]);

  const completeMessage = useCallback((status = 'complete') => {
    // console.log("Complete Message: ", status)
    setMessage((prevMessage) => {
      const lastMessage = prevMessage[prevMessage.length - 1];
      // only change the status if the last message is not complete and not error
      if (lastMessage.status === 'complete' || lastMessage.status === 'error') {
        return prevMessage;
      }
      return [
        ...prevMessage.slice(0, -1),
        { ...lastMessage, status: status }
      ];
    });
  }, [])

  const generateMockResponse = useCallback((content) => {
    // console.log("Generate Mock Response: ", content);
    setMessage((message) => {
      const lastMessage = message[message.length - 1];
      let newMessage = {...lastMessage};
      if (lastMessage.status === 'loading' || lastMessage.status === 'incomplete') {
        newMessage = {
          ...newMessage,
          content: (lastMessage.content || '') + content,
          status: 'incomplete'
        }
      }
      return [ ...message.slice(0, -1), newMessage ]
    })
  }, []);

  const SettingsToggle = () => {
    if (!styleState.isMobile) return null;
    return (
      <Button
        icon={<IconSetting />}
        style={{
          position: 'absolute',
          left: showSettings ? -10 : -20,
          top: '50%',
          transform: 'translateY(-50%)',
          zIndex: 1000,
          width: 40,
          height: 40,
          borderRadius: '0 20px 20px 0',
          padding: 0,
          boxShadow: '2px 0 8px rgba(0, 0, 0, 0.15)',
        }}
        onClick={() => setShowSettings(!showSettings)}
        theme="solid"
        type="primary"
      />
    );
  };

  function CustomInputRender(props) {
    const { detailProps } = props;
    const { clearContextNode, uploadNode, inputNode, sendNode, onClick } = detailProps;

    return <div style={{margin: '8px 16px', display: 'flex', flexDirection:'row',
      alignItems: 'flex-end', borderRadius: 16,padding: 10, border: '1px solid var(--semi-color-border)'}}
                onClick={onClick}
    >
      {/*{uploadNode}*/}
      {inputNode}
      {sendNode}
    </div>
  }

  const renderInputArea = useCallback((props) => {
    return (<CustomInputRender {...props} />)
  }, []);

  return (
    <Layout style={{height: '100%'}}>
      {(showSettings || !styleState.isMobile) && (
        <Layout.Sider style={{ display: styleState.isMobile ? 'block' : 'initial' }}>
          <Card style={commonOuterStyle}>
            <div style={{ marginTop: 10 }}>
              <Typography.Text strong>{t('分组')}：</Typography.Text>
            </div>
            <Select
              placeholder={t('请选择分组')}
              name='group'
              required
              selection
              onChange={(value) => {
                handleInputChange('group', value);
              }}
              value={inputs.group}
              autoComplete='new-password'
              optionList={groups}
              renderOptionItem={renderGroupOption}
              style={{ width: '100%' }}
            />
            <div style={{ marginTop: 10 }}>
              <Typography.Text strong>{t('模型')}：</Typography.Text>
            </div>
            <Select
              placeholder={t('请选择模型')}
              name='model'
              required
              selection
              searchPosition='dropdown'
              filter
              onChange={(value) => {
                handleInputChange('model', value);
              }}
              value={inputs.model}
              autoComplete='new-password'
              optionList={models}
            />
            <div style={{ marginTop: 10 }}>
              <Typography.Text strong>Temperature：</Typography.Text>
            </div>
            <Slider
              step={0.1}
              min={0.1}
              max={1}
              value={inputs.temperature}
              onChange={(value) => {
                handleInputChange('temperature', value);
              }}
            />
            <div style={{ marginTop: 10 }}>
              <Typography.Text strong>MaxTokens：</Typography.Text>
            </div>
            <Input
              placeholder='MaxTokens'
              name='max_tokens'
              required
              autoComplete='new-password'
              defaultValue={0}
              value={inputs.max_tokens}
              onChange={(value) => {
                handleInputChange('max_tokens', value);
              }}
            />

            <div style={{ marginTop: 10 }}>
              <Typography.Text strong>System：</Typography.Text>
            </div>
            <TextArea
              placeholder='System Prompt'
              name='system'
              required
              autoComplete='new-password'
              autosize
              defaultValue={systemPrompt}
              // value={systemPrompt}
              onChange={(value) => {
                setSystemPrompt(value);
              }}
            />

          </Card>
        </Layout.Sider>
      )}
      <Layout.Content>
        <div style={{height: '100%', position: 'relative'}}>
          <SettingsToggle />
          <Chat
            chatBoxRenderConfig={{
              renderChatBoxAction: () => {
                return <div></div>
              }
            }}
            renderInputArea={renderInputArea}
            roleConfig={roleInfo}
            style={commonOuterStyle}
            chats={message}
            onMessageSend={onMessageSend}
            showClearContext
            onClear={() => {
              setMessage([]);
            }}
          />
        </div>
      </Layout.Content>
    </Layout>
  );
};

export default Playground;
