import React, { useEffect, useRef, useState } from 'react';
import { initVChartSemiTheme } from '@visactor/vchart-semi-theme';

import { Button, Col, Form, Layout, Row, Spin } from '@douyinfe/semi-ui';
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

const Detail = (props) => {
  const formRef = useRef();
  let now = new Date();
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
      text: '模型调用次数占比',
      subtext: `总计：${renderNumber(times)}`,
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
      text: '模型消耗分布',
      subtext: `总计：${renderQuota(consumeQuota, 2)}`,
    },
    bar: {
      // The state style of bar
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
          // sort by value
          array.sort((a, b) => b.value - a.value);
          // add $
          let sum = 0;
          for (let i = 0; i < array.length; i++) {
            sum += parseFloat(array[i].value);
            array[i].value = renderQuotaNumberWithDigit(
              parseFloat(array[i].value),
              4,
            );
          }
          // add to first
          array.unshift({
            key: '总计',
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

    // 首先收集所有唯一的模型名称
    data.forEach(item => uniqueModels.add(item.model_name));
    
    // 为每个唯一的模型生成或获取颜色
    const newModelColors = {};
    Array.from(uniqueModels).forEach((modelName) => {
      // 优先使用 modelColorMap 中的颜色，然后是已存在的颜色，最后使用新的颜色生成函数
      newModelColors[modelName] = modelColorMap[modelName] || 
        modelColors[modelName] || 
        modelToColor(modelName);  // 使用新的颜色生成函数替代 stringToColor
    });
    setModelColors(newModelColors);

    for (let i = 0; i < data.length; i++) {
      const item = data[i];
      totalQuota += item.quota;
      totalTimes += item.count;
      // 合并model_name
      let pieItem = newPieData.find((it) => it.type === item.model_name);
      if (pieItem) {
        pieItem.value += item.count;
      } else {
        newPieData.push({
          type: item.model_name,
          value: item.count,
        });
      }
      // 合并created_at和model_name 为 lineData
      let createTime = timestamp2string1(
        item.created_at,
        dataExportDefaultTime,
      );
      let lineItem = newLineData.find(
        (it) => it.Time === createTime && it.Model === item.model_name,
      );
      if (lineItem) {
        lineItem.Usage += parseFloat(getQuotaWithUnit(item.quota));
      } else {
        newLineData.push({
          Time: createTime,
          Model: item.model_name,
          Usage: parseFloat(getQuotaWithUnit(item.quota)),
        });
      }
    }

    // sort by count
    newPieData.sort((a, b) => b.value - a.value);

    // 更新图表配置和数据
    setSpecPie(prev => ({
      ...prev,
      data: [{ id: 'id0', values: newPieData }],
      title: {
        ...prev.title,
        subtext: `总计：${renderNumber(totalTimes)}`
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
        subtext: `总计：${renderQuota(totalQuota, 2)}`
      },
      color: {
        specified: newModelColors
      }
    }));
    
    setPieData(newPieData);
    setLineData(newLineData);
    setConsumeQuota(totalQuota);
    setTimes(totalTimes);
  };

  useEffect(() => {
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
          <h3>数据看板</h3>
        </Layout.Header>
        <Layout.Content>
          <Form ref={formRef} layout='horizontal' style={{ marginTop: 10 }}>
            <>
              <Form.DatePicker
                field='start_timestamp'
                label='起始时间'
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
                label='结束时间'
                style={{ width: 272 }}
                initValue={end_timestamp}
                value={end_timestamp}
                type='dateTime'
                name='end_timestamp'
                onChange={(value) => handleInputChange(value, 'end_timestamp')}
              />
              <Form.Select
                field='data_export_default_time'
                label='时间粒度'
                style={{ width: 176 }}
                initValue={dataExportDefaultTime}
                placeholder={'时间粒度'}
                name='data_export_default_time'
                optionList={[
                  { label: '小时', value: 'hour' },
                  { label: '天', value: 'day' },
                  { label: '周', value: 'week' },
                ]}
                onChange={(value) =>
                  handleInputChange(value, 'data_export_default_time')
                }
              ></Form.Select>
              {isAdminUser && (
                <>
                  <Form.Input
                    field='username'
                    label='用户名称'
                    style={{ width: 176 }}
                    value={username}
                    placeholder={'可选值'}
                    name='username'
                    onChange={(value) => handleInputChange(value, 'username')}
                  />
                </>
              )}
              <Form.Section>
                <Button
                  label='查询'
                  type='primary'
                  htmlType='submit'
                  className='btn-margin-right'
                  onClick={refresh}
                  loading={loading}
                >
                  查询
                </Button>
              </Form.Section>
            </>
          </Form>
          <Spin spinning={loading}>
            <div style={{ height: 500 }}>
              <VChart 
                spec={spec_pie}
                option={{ mode: "desktop-browser" }}
              />
            </div>
            <div style={{ height: 500 }}>
              <VChart 
                spec={spec_line}
                option={{ mode: "desktop-browser" }}
              />
            </div>
          </Spin>
        </Layout.Content>
      </Layout>
    </>
  );
};

export default Detail;
