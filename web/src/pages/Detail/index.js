import React, { useContext, useEffect, useRef, useState } from 'react';
import { initVChartSemiTheme } from '@visactor/vchart-semi-theme';

import { Button, Card, Col, Descriptions, Form, Layout, Row, Spin, Tabs } from '@douyinfe/semi-ui';
import { VChart } from "@visactor/react-vchart";
import {
  API,
  isAdmin,
  showError,
  timestamp2string,
  timestamp2string1,
} from '../../helpers';
import {
  getQuotaWithUnit,
  modelColorMap,
  renderNumber,
  renderQuota,
  renderQuotaNumberWithDigit,
  stringToColor,
  modelToColor,
} from '../../helpers/render';
import { UserContext } from '../../context/User/index.js';
import { StyleContext } from '../../context/Style/index.js';
import { useTranslation } from 'react-i18next';

const Detail = (props) => {
  const { t } = useTranslation();
  const formRef = useRef();
  let now = new Date();
  const [userState, userDispatch] = useContext(UserContext);
  const [styleState, styleDispatch] = useContext(StyleContext);
  const [inputs, setInputs] = useState({
    username: '',
    token_name: '',
    model_name: '',
    start_timestamp:
      localStorage.getItem('data_export_default_time') === 'hour'
        ? timestamp2string(now.getTime() / 1000 - 86400)
        : localStorage.getItem('data_export_default_time') === 'week'
          ? timestamp2string(now.getTime() / 1000 - 86400 * 30)
          : timestamp2string(now.getTime() / 1000 - 86400 * 7),
    end_timestamp: timestamp2string(now.getTime() / 1000 + 3600),
    channel: '',
    data_export_default_time: '',
  });
  const { username, model_name, start_timestamp, end_timestamp, channel } =
    inputs;
  const isAdminUser = isAdmin();
  const initialized = useRef(false);
  const [loading, setLoading] = useState(false);
  const [quotaData, setQuotaData] = useState([]);
  const [consumeQuota, setConsumeQuota] = useState(0);
  const [consumeTokens, setConsumeTokens] = useState(0);
  const [times, setTimes] = useState(0);
  const [dataExportDefaultTime, setDataExportDefaultTime] = useState(
    localStorage.getItem('data_export_default_time') || 'hour',
  );
  const [pieData, setPieData] = useState([{ type: 'null', value: '0' }]);
  const [lineData, setLineData] = useState([]);
  const [spec_pie, setSpecPie] = useState({
    type: 'pie',
    data: [{
      id: 'id0',
      values: pieData
    }],
    outerRadius: 0.8,
    innerRadius: 0.5,
    padAngle: 0.6,
    valueField: 'value',
    categoryField: 'type',
    pie: {
      style: {
        cornerRadius: 10,
      },
      state: {
        hover: {
          outerRadius: 0.85,
          stroke: '#000',
          lineWidth: 1,
        },
        selected: {
          outerRadius: 0.85,
          stroke: '#000',
          lineWidth: 1,
        },
      },
    },
    title: {
      visible: true,
      text: t('模型调用次数占比'),
      subtext: `${t('总计')}：${renderNumber(times)}`,
    },
    legends: {
      visible: true,
      orient: 'left',
    },
    label: {
      visible: true,
    },
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['type'],
            value: (datum) => renderNumber(datum['value']),
          },
        ],
      },
    },
    color: {
      specified: modelColorMap,
    },
  });
  const [spec_line, setSpecLine] = useState({
    type: 'bar',
    data: [{
      id: 'barData',
      values: lineData
    }],
    xField: 'Time',
    yField: 'Usage',
    seriesField: 'Model',
    stack: true,
    legends: {
      visible: true,
      selectMode: 'single',
    },
    title: {
      visible: true,
      text: t('模型消耗分布'),
      subtext: `${t('总计')}：${renderQuota(consumeQuota, 2)}`,
    },
    bar: {
      state: {
        hover: {
          stroke: '#000',
          lineWidth: 1,
        },
      },
    },
    tooltip: {
      mark: {
        content: [
          {
            key: (datum) => datum['Model'],
            value: (datum) =>
              renderQuotaNumberWithDigit(parseFloat(datum['Usage']), 4),
          },
        ],
      },
      dimension: {
        content: [
          {
            key: (datum) => datum['Model'],
            value: (datum) => datum['Usage'],
          },
        ],
        updateContent: (array) => {
          array.sort((a, b) => b.value - a.value);
          let sum = 0;
          for (let i = 0; i < array.length; i++) {
            sum += parseFloat(array[i].value);
            array[i].value = renderQuotaNumberWithDigit(
              parseFloat(array[i].value),
              4,
            );
          }
          array.unshift({
            key: t('总计'),
            value: renderQuotaNumberWithDigit(sum, 4),
          });
          return array;
        },
      },
    },
    color: {
      specified: modelColorMap,
    },
  });

  // 添加一个新的状态来存储模型-颜色映射
  const [modelColors, setModelColors] = useState({});

  const handleInputChange = (value, name) => {
    if (name === 'data_export_default_time') {
      setDataExportDefaultTime(value);
      return;
    }
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadQuotaData = async () => {
    setLoading(true);
    try {
      let url = '';
      let localStartTimestamp = Date.parse(start_timestamp) / 1000;
      let localEndTimestamp = Date.parse(end_timestamp) / 1000;
      if (isAdminUser) {
        url = `/api/data/?username=${username}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&default_time=${dataExportDefaultTime}`;
      } else {
        url = `/api/data/self/?start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&default_time=${dataExportDefaultTime}`;
      }
      const res = await API.get(url);
      const { success, message, data } = res.data;
      if (success) {
        setQuotaData(data);
        if (data.length === 0) {
          data.push({
            count: 0,
            model_name: '无数据',
            quota: 0,
            created_at: now.getTime() / 1000,
          });
        }
        // 根据dataExportDefaultTime重制时间粒度
        let timeGranularity = 3600;
        if (dataExportDefaultTime === 'day') {
          timeGranularity = 86400;
        } else if (dataExportDefaultTime === 'week') {
          timeGranularity = 604800;
        }
        // sort created_at
        data.sort((a, b) => a.created_at - b.created_at);
        data.forEach((item) => {
          item['created_at'] =
            Math.floor(item['created_at'] / timeGranularity) * timeGranularity;
        });
        updateChartData(data);
      } else {
        showError(message);
      }
    } finally {
      setLoading(false);
    }
  };

  const refresh = async () => {
    await loadQuotaData();
  };

  const initChart = async () => {
    await loadQuotaData();
  };

  const updateChartData = (data) => {
    let newPieData = [];
    let newLineData = [];
    let totalQuota = 0;
    let totalTimes = 0;
    let uniqueModels = new Set();
    let totalTokens = 0;

    // 收集所有唯一的模型名称和时间点
    let uniqueTimes = new Set();
    data.forEach(item => {
      uniqueModels.add(item.model_name);
      uniqueTimes.add(timestamp2string1(item.created_at, dataExportDefaultTime));
      totalTokens += item.token_used;
    });
    
    // 处理颜色映射
    const newModelColors = {};
    Array.from(uniqueModels).forEach((modelName) => {
      newModelColors[modelName] = modelColorMap[modelName] || 
        modelColors[modelName] || 
        modelToColor(modelName);
    });
    setModelColors(newModelColors);

    // 处理饼图数据
    for (let item of data) {
      totalQuota += item.quota;
      totalTimes += item.count;
      
      let pieItem = newPieData.find((it) => it.type === item.model_name);
      if (pieItem) {
        pieItem.value += item.count;
      } else {
        newPieData.push({
          type: item.model_name,
          value: item.count,
        });
      }
    }

    // 处理柱状图数据
    let timePoints = Array.from(uniqueTimes);
    if (timePoints.length < 7) {
      // 根据时间粒度生成合适的时间点
      const generateTimePoints = () => {
        let lastTime = Math.max(...data.map(item => item.created_at));
        let points = [];
        let interval = dataExportDefaultTime === 'hour' ? 3600 
                      : dataExportDefaultTime === 'day' ? 86400 
                      : 604800;

        for (let i = 0; i < 7; i++) {
          points.push(timestamp2string1(lastTime - (i * interval), dataExportDefaultTime));
        }
        return points.reverse();
      };

      timePoints = generateTimePoints();
    }

    // 为每个时间点和模型生成数据
    timePoints.forEach(time => {
      Array.from(uniqueModels).forEach(model => {
        let existingData = data.find(item => 
          timestamp2string1(item.created_at, dataExportDefaultTime) === time && 
          item.model_name === model
        );

        newLineData.push({
          Time: time,
          Model: model,
          Usage: existingData ? parseFloat(getQuotaWithUnit(existingData.quota)) : 0
        });
      });
    });

    // 排序
    newPieData.sort((a, b) => b.value - a.value);
    newLineData.sort((a, b) => a.Time.localeCompare(b.Time));

    // 更新图表配置和数据
    setSpecPie(prev => ({
      ...prev,
      data: [{ id: 'id0', values: newPieData }],
      title: {
        ...prev.title,
        subtext: `${t('总计')}：${renderNumber(totalTimes)}`
      },
      color: {
        specified: newModelColors
      }
    }));

    setSpecLine(prev => ({
      ...prev,
      data: [{ id: 'barData', values: newLineData }],
      title: {
        ...prev.title,
        subtext: `${t('总计')}：${renderQuota(totalQuota, 2)}`
      },
      color: {
        specified: newModelColors
      }
    }));
    
    setPieData(newPieData);
    setLineData(newLineData);
    setConsumeQuota(totalQuota);
    setTimes(totalTimes);
    setConsumeTokens(totalTokens);
  };

  const getUserData = async () => {
    let res = await API.get(`/api/user/self`);
    const {success, message, data} = res.data;
    if (success) {
      userDispatch({type: 'login', payload: data});
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    getUserData()
    if (!initialized.current) {
      initVChartSemiTheme({
        isWatchingThemeSwitch: true,
      });
      initialized.current = true;
      initChart();
    }
  }, []);

  return (
    <>
      <Layout>
        <Layout.Header>
          <h3>{t('数据看板')}</h3>
        </Layout.Header>
        <Layout.Content>
          <Form ref={formRef} layout='horizontal' style={{ marginTop: 10 }}>
            <>
              <Form.DatePicker
                field='start_timestamp'
                label={t('起始时间')}
                style={{ width: 272 }}
                initValue={start_timestamp}
                value={start_timestamp}
                type='dateTime'
                name='start_timestamp'
                onChange={(value) =>
                  handleInputChange(value, 'start_timestamp')
                }
              />
              <Form.DatePicker
                field='end_timestamp'
                fluid
                label={t('结束时间')}
                style={{ width: 272 }}
                initValue={end_timestamp}
                value={end_timestamp}
                type='dateTime'
                name='end_timestamp'
                onChange={(value) => handleInputChange(value, 'end_timestamp')}
              />
              <Form.Select
                field='data_export_default_time'
                label={t('时间粒度')}
                style={{ width: 176 }}
                initValue={dataExportDefaultTime}
                placeholder={t('时间粒度')}
                name='data_export_default_time'
                optionList={[
                  { label: t('小时'), value: 'hour' },
                  { label: t('天'), value: 'day' },
                  { label: t('周'), value: 'week' },
                ]}
                onChange={(value) =>
                  handleInputChange(value, 'data_export_default_time')
                }
              ></Form.Select>
              {isAdminUser && (
                <>
                  <Form.Input
                    field='username'
                    label={t('用户名称')}
                    style={{ width: 176 }}
                    value={username}
                    placeholder={t('可选值')}
                    name='username'
                    onChange={(value) => handleInputChange(value, 'username')}
                  />
                </>
              )}
              <Button
                label={t('查询')}
                type='primary'
                htmlType='submit'
                className='btn-margin-right'
                onClick={refresh}
                loading={loading}
                style={{ marginTop: 24 }}
              >
                {t('查询')}
              </Button>
              <Form.Section>
              </Form.Section>
            </>
          </Form>
          <Spin spinning={loading}>
            <Row gutter={{ xs: 16, sm: 16, md: 16, lg: 24, xl: 24, xxl: 24 }} style={{marginTop: 20}} type="flex" justify="space-between">
              <Col span={styleState.isMobile?24:8}>
                <Card className='panel-desc-card'>
                  <Descriptions row size="small">
                    <Descriptions.Item itemKey={t('当前余额')}>
                      {renderQuota(userState?.user?.quota)}
                    </Descriptions.Item>
                    <Descriptions.Item itemKey={t('历史消耗')}>
                      {renderQuota(userState?.user?.used_quota)}
                    </Descriptions.Item>
                    <Descriptions.Item itemKey={t('请求次数')}>
                      {userState.user?.request_count}
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={styleState.isMobile?24:8}>
                <Card>
                  <Descriptions row size="small">
                    <Descriptions.Item itemKey={t('统计额度')}>
                      {renderQuota(consumeQuota)}
                    </Descriptions.Item>
                    <Descriptions.Item itemKey={t('统计Tokens')}>
                      {consumeTokens}
                    </Descriptions.Item>
                    <Descriptions.Item itemKey={t('统计次数')}>
                      {times}
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col span={styleState.isMobile ? 24 : 8}>
                <Card>
                  <Descriptions row size='small'>
                    <Descriptions.Item itemKey={t('平均RPM')}>
                      {(times /
                        ((Date.parse(end_timestamp) -
                          Date.parse(start_timestamp)) /
                          60000)).toFixed(3)}
                    </Descriptions.Item>
                    <Descriptions.Item itemKey={t('平均TPM')}>
                      {(consumeTokens /
                        ((Date.parse(end_timestamp) -
                          Date.parse(start_timestamp)) /
                          60000)).toFixed(3)}
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
            </Row>
            <Card style={{marginTop: 20}}>
              <Tabs type="line" defaultActiveKey="1">
                <Tabs.TabPane tab={t('消耗分布')} itemKey="1">
                  <div style={{ height: 500 }}>
                    <VChart
                      spec={spec_line}
                      option={{ mode: "desktop-browser" }}
                    />
                  </div>
                </Tabs.TabPane>
                <Tabs.TabPane tab={t('调用次数分布')} itemKey="2">
                  <div style={{ height: 500 }}>
                    <VChart
                      spec={spec_pie}
                      option={{ mode: "desktop-browser" }}
                    />
                  </div>
                </Tabs.TabPane>

              </Tabs>
            </Card>
          </Spin>
        </Layout.Content>
      </Layout>
    </>
  );
};

export default Detail;
