import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  API,
  isMobile,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';
import { renderGroupOption, renderQuotaWithPrompt } from '../../helpers/render';
import {
  AutoComplete,
  Banner,
  Button,
  Checkbox,
  DatePicker,
  Input,
  Select,
  SideSheet,
  Space,
  Spin,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import { Divider } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';

const EditToken = (props) => {
  const [isEdit, setIsEdit] = useState(false);
  const [loading, setLoading] = useState(isEdit);
  const originInputs = {
    name: '',
    remain_quota: isEdit ? 0 : 500000,
    expired_time: -1,
    unlimited_quota: false,
    model_limits_enabled: false,
    model_limits: [],
    allow_ips: '',
    group: '',
  };
  const [inputs, setInputs] = useState(originInputs);
  const {
    name,
    remain_quota,
    expired_time,
    unlimited_quota,
    model_limits_enabled,
    model_limits,
    allow_ips,
    group,
  } = inputs;
  // const [visible, setVisible] = useState(false);
  const [models, setModels] = useState([]);
  const [groups, setGroups] = useState([]);
  const navigate = useNavigate();
  const { t } = useTranslation();
  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };
  const handleCancel = () => {
    props.handleClose();
  };
  const setExpiredTime = (month, day, hour, minute) => {
    let now = new Date();
    let timestamp = now.getTime() / 1000;
    let seconds = month * 30 * 24 * 60 * 60;
    seconds += day * 24 * 60 * 60;
    seconds += hour * 60 * 60;
    seconds += minute * 60;
    if (seconds !== 0) {
      timestamp += seconds;
      setInputs({ ...inputs, expired_time: timestamp2string(timestamp) });
    } else {
      setInputs({ ...inputs, expired_time: -1 });
    }
  };

  const setUnlimitedQuota = () => {
    setInputs({ ...inputs, unlimited_quota: !unlimited_quota });
  };

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
        ratio: info.ratio,
      }));
      setGroups(localGroupOptions);
    } else {
      showError(t(message));
    }
  };

  const loadToken = async () => {
    setLoading(true);
    let res = await API.get(`/api/token/${props.editingToken.id}`);
    const { success, message, data } = res.data;
    if (success) {
      if (data.expired_time !== -1) {
        data.expired_time = timestamp2string(data.expired_time);
      }
      if (data.model_limits !== '') {
        data.model_limits = data.model_limits.split(',');
      } else {
        data.model_limits = [];
      }
      setInputs(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };
  useEffect(() => {
    setIsEdit(props.editingToken.id !== undefined);
  }, [props.editingToken.id]);

  useEffect(() => {
    if (!isEdit) {
      setInputs(originInputs);
    } else {
      loadToken().then(() => {
        // console.log(inputs);
      });
    }
    loadModels();
    loadGroups();
  }, [isEdit]);

  // 新增 state 变量 tokenCount 来记录用户想要创建的令牌数量，默认为 1
  const [tokenCount, setTokenCount] = useState(1);

  // 新增处理 tokenCount 变化的函数
  const handleTokenCountChange = (value) => {
    // 确保用户输入的是正整数
    const count = parseInt(value, 10);
    if (!isNaN(count) && count > 0) {
      setTokenCount(count);
    }
  };

  // 生成一个随机的四位字母数字字符串
  const generateRandomSuffix = () => {
    const characters =
      'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < 6; i++) {
      result += characters.charAt(
        Math.floor(Math.random() * characters.length),
      );
    }
    return result;
  };

  const submit = async () => {
    setLoading(true);
    if (isEdit) {
      // 编辑令牌的逻辑保持不变
      let localInputs = { ...inputs };
      localInputs.remain_quota = parseInt(localInputs.remain_quota);
      if (localInputs.expired_time !== -1) {
        let time = Date.parse(localInputs.expired_time);
        if (isNaN(time)) {
          showError(t('过期时间格式错误！'));
          setLoading(false);
          return;
        }
        localInputs.expired_time = Math.ceil(time / 1000);
      }
      localInputs.model_limits = localInputs.model_limits.join(',');
      let res = await API.put(`/api/token/`, {
        ...localInputs,
        id: parseInt(props.editingToken.id),
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('令牌更新成功！'));
        props.refresh();
        props.handleClose();
      } else {
        showError(t(message));
      }
    } else {
      // 处理新增多个令牌的情况
      let successCount = 0; // 记录成功创建的令牌数量
      for (let i = 0; i < tokenCount; i++) {
        let localInputs = { ...inputs };
        if (i !== 0) {
          // 如果用户想要创建多个令牌，则给每个令牌一个序号后缀
          localInputs.name = `${inputs.name}-${generateRandomSuffix()}`;
        }
        localInputs.remain_quota = parseInt(localInputs.remain_quota);

        if (localInputs.expired_time !== -1) {
          let time = Date.parse(localInputs.expired_time);
          if (isNaN(time)) {
            showError(t('过期时间格式错误！'));
            setLoading(false);
            break;
          }
          localInputs.expired_time = Math.ceil(time / 1000);
        }
        localInputs.model_limits = localInputs.model_limits.join(',');
        let res = await API.post(`/api/token/`, localInputs);
        const { success, message } = res.data;

        if (success) {
          successCount++;
        } else {
          showError(t(message));
          break; // 如果创建失败，终止循环
        }
      }

      if (successCount > 0) {
        showSuccess(t('令牌创建成功，请在列表页面点击复制获取令牌！'));
        props.refresh();
        props.handleClose();
      }
    }
    setLoading(false);
    setInputs(originInputs); // 重置表单
    setTokenCount(1); // 重置数量为默认值
  };

  return (
    <>
      <SideSheet
        placement={isEdit ? 'right' : 'left'}
        title={
          <Title level={3}>
            {isEdit ? t('更新令牌信息') : t('创建新的令牌')}
          </Title>
        }
        headerStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        bodyStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        visible={props.visiable}
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button theme='solid' size={'large'} onClick={submit}>
                {t('提交')}
              </Button>
              <Button
                theme='solid'
                size={'large'}
                type={'tertiary'}
                onClick={handleCancel}
              >
                {t('取消')}
              </Button>
            </Space>
          </div>
        }
        closeIcon={null}
        onCancel={() => handleCancel()}
        width={isMobile() ? '100%' : 600}
      >
        <Spin spinning={loading}>
          <Input
            style={{ marginTop: 20 }}
            label={t('名称')}
            name='name'
            placeholder={t('请输入名称')}
            onChange={(value) => handleInputChange('name', value)}
            value={name}
            autoComplete='new-password'
            required={!isEdit}
          />
          <Divider />
          <DatePicker
            label={t('过期时间')}
            name='expired_time'
            placeholder={t('请选择过期时间')}
            onChange={(value) => handleInputChange('expired_time', value)}
            value={expired_time}
            autoComplete='new-password'
            type='dateTime'
          />
          <div style={{ marginTop: 20 }}>
            <Space>
              <Button
                type={'tertiary'}
                onClick={() => {
                  setExpiredTime(0, 0, 0, 0);
                }}
              >
                {t('永不过期')}
              </Button>
              <Button
                type={'tertiary'}
                onClick={() => {
                  setExpiredTime(0, 0, 1, 0);
                }}
              >
                {t('一小时')}
              </Button>
              <Button
                type={'tertiary'}
                onClick={() => {
                  setExpiredTime(1, 0, 0, 0);
                }}
              >
                {t('一个月')}
              </Button>
              <Button
                type={'tertiary'}
                onClick={() => {
                  setExpiredTime(0, 1, 0, 0);
                }}
              >
                {t('一天')}
              </Button>
            </Space>
          </div>

          <Divider />
          <Banner
            type={'warning'}
            description={t(
              '注意，令牌的额度仅用于限制令牌本身的最大额度使用量，实际的使用受到账户的剩余额度限制。',
            )}
          ></Banner>
          <div style={{ marginTop: 20 }}>
            <Typography.Text>{`${t('额度')}${renderQuotaWithPrompt(remain_quota)}`}</Typography.Text>
          </div>
          <AutoComplete
            style={{ marginTop: 8 }}
            name='remain_quota'
            placeholder={t('请输入额度')}
            onChange={(value) => handleInputChange('remain_quota', value)}
            value={remain_quota}
            autoComplete='new-password'
            type='number'
            // position={'top'}
            data={[
              { value: 500000, label: '1$' },
              { value: 5000000, label: '10$' },
              { value: 25000000, label: '50$' },
              { value: 50000000, label: '100$' },
              { value: 250000000, label: '500$' },
              { value: 500000000, label: '1000$' },
            ]}
            disabled={unlimited_quota}
          />

          {!isEdit && (
            <>
              <div style={{ marginTop: 20 }}>
                <Typography.Text>{t('新建数量')}</Typography.Text>
              </div>
              <AutoComplete
                style={{ marginTop: 8 }}
                label={t('数量')}
                placeholder={t('请选择或输入创建令牌的数量')}
                onChange={(value) => handleTokenCountChange(value)}
                onSelect={(value) => handleTokenCountChange(value)}
                value={tokenCount.toString()}
                autoComplete='off'
                type='number'
                data={[
                  { value: 10, label: t('10个') },
                  { value: 20, label: t('20个') },
                  { value: 30, label: t('30个') },
                  { value: 100, label: t('100个') },
                ]}
                disabled={unlimited_quota}
              />
            </>
          )}

          <div>
            <Button
              style={{ marginTop: 8 }}
              type={'warning'}
              onClick={() => {
                setUnlimitedQuota();
              }}
            >
              {unlimited_quota ? t('取消无限额度') : t('设为无限额度')}
            </Button>
          </div>
          <Divider />
          <div style={{ marginTop: 10 }}>
            <Typography.Text>
              {t('IP白名单（请勿过度信任此功能）')}
            </Typography.Text>
          </div>
          <TextArea
            label={t('IP白名单')}
            name='allow_ips'
            placeholder={t('允许的IP，一行一个，不填写则不限制')}
            onChange={(value) => {
              handleInputChange('allow_ips', value);
            }}
            value={inputs.allow_ips}
            style={{ fontFamily: 'JetBrains Mono, Consolas' }}
          />
          <div style={{ marginTop: 10, display: 'flex' }}>
            <Space>
              <Checkbox
                name='model_limits_enabled'
                checked={model_limits_enabled}
                onChange={(e) =>
                  handleInputChange('model_limits_enabled', e.target.checked)
                }
              >
                {t('启用模型限制（非必要，不建议启用）')}
              </Checkbox>
            </Space>
          </div>

          <Select
            style={{ marginTop: 8 }}
            placeholder={t('请选择该渠道所支持的模型')}
            name='models'
            required
            multiple
            selection
            onChange={(value) => {
              handleInputChange('model_limits', value);
            }}
            value={inputs.model_limits}
            autoComplete='new-password'
            optionList={models}
            disabled={!model_limits_enabled}
          />
          <div style={{ marginTop: 10 }}>
            <Typography.Text>{t('令牌分组，默认为用户的分组')}</Typography.Text>
          </div>
          {groups.length > 0 ? (
            <Select
              style={{ marginTop: 8 }}
              placeholder={t('令牌分组，默认为用户的分组')}
              name='gruop'
              required
              selection
              onChange={(value) => {
                handleInputChange('group', value);
              }}
              position={'topLeft'}
              renderOptionItem={renderGroupOption}
              value={inputs.group}
              autoComplete='new-password'
              optionList={groups}
            />
          ) : (
            <Select
              style={{ marginTop: 8 }}
              placeholder={t('管理员未设置用户可选分组')}
              name='gruop'
              disabled={true}
            />
          )}
        </Spin>
      </SideSheet>
    </>
  );
};

export default EditToken;
