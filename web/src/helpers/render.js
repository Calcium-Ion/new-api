import i18next from 'i18next';
import { Modal, Tag, Typography } from '@douyinfe/semi-ui';
import { copy, isMobile, showSuccess } from './utils.js';

export function renderText(text, limit) {
  if (text.length > limit) {
    return text.slice(0, limit - 3) + '...';
  }
  return text;
}

/**
 * Render group tags based on the input group string
 * @param {string} group - The input group string
 * @returns {JSX.Element} - The rendered group tags
 */
export function renderGroup(group) {
  if (group === '') {
    return (
      <Tag size='large' key='default' color='orange'>
        {i18next.t('用户分组')}
      </Tag>
    );
  }

  const tagColors = {
    vip: 'yellow',
    pro: 'yellow',
    svip: 'red',
    premium: 'red',
  };

  const groups = group.split(',').sort();

  return (
    <span key={group}>
      {groups.map((group) => (
        <Tag
          size='large'
          color={tagColors[group] || stringToColor(group)}
          key={group}
          onClick={async (event) => {
            event.stopPropagation();
            if (await copy(group)) {
              showSuccess(i18next.t('已复制：') + group);
            } else {
              Modal.error({
                title: t('无法复制到剪贴板，请手动复制'),
                content: group,
              });
            }
          }}
        >
          {group}
        </Tag>
      ))}
    </span>
  );
}

export function renderRatio(ratio) {
  let color = 'green';
  if (ratio > 5) {
    color = 'red';
  } else if (ratio > 3) {
    color = 'orange';
  } else if (ratio > 1) {
    color = 'blue';
  }
  return (
    <Tag color={color}>
      {ratio}x {i18next.t('倍率')}
    </Tag>
  );
}

const measureTextWidth = (
  text,
  style = {
    fontSize: '14px',
    fontFamily:
      '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
  },
  containerWidth,
) => {
  const span = document.createElement('span');

  span.style.visibility = 'hidden';
  span.style.position = 'absolute';
  span.style.whiteSpace = 'nowrap';
  span.style.fontSize = style.fontSize;
  span.style.fontFamily = style.fontFamily;

  span.textContent = text;

  document.body.appendChild(span);
  const width = span.offsetWidth;

  document.body.removeChild(span);

  return width;
};

export function truncateText(text, maxWidth = 200) {
  if (!isMobile()) {
    return text;
  }
  if (!text) return text;

  try {
    // Handle percentage-based maxWidth
    let actualMaxWidth = maxWidth;
    if (typeof maxWidth === 'string' && maxWidth.endsWith('%')) {
      const percentage = parseFloat(maxWidth) / 100;
      // Use window width as fallback container width
      actualMaxWidth = window.innerWidth * percentage;
    }

    const width = measureTextWidth(text);
    if (width <= actualMaxWidth) return text;

    let left = 0;
    let right = text.length;
    let result = text;

    while (left <= right) {
      const mid = Math.floor((left + right) / 2);
      const truncated = text.slice(0, mid) + '...';
      const currentWidth = measureTextWidth(truncated);

      if (currentWidth <= actualMaxWidth) {
        result = truncated;
        left = mid + 1;
      } else {
        right = mid - 1;
      }
    }

    return result;
  } catch (error) {
    console.warn(
      'Text measurement failed, falling back to character count',
      error,
    );
    if (text.length > 20) {
      return text.slice(0, 17) + '...';
    }
    return text;
  }
}

export const renderGroupOption = (item) => {
  const {
    disabled,
    selected,
    label,
    value,
    focused,
    className,
    style,
    onMouseEnter,
    onClick,
    empty,
    emptyContent,
    ...rest
  } = item;

  const baseStyle = {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '8px 16px',
    cursor: disabled ? 'not-allowed' : 'pointer',
    backgroundColor: focused ? 'var(--semi-color-fill-0)' : 'transparent',
    opacity: disabled ? 0.5 : 1,
    ...(selected && {
      backgroundColor: 'var(--semi-color-primary-light-default)',
    }),
    '&:hover': {
      backgroundColor: !disabled && 'var(--semi-color-fill-1)',
    },
  };

  const handleClick = () => {
    if (!disabled && onClick) {
      onClick();
    }
  };

  const handleMouseEnter = (e) => {
    if (!disabled && onMouseEnter) {
      onMouseEnter(e);
    }
  };

  return (
    <div
      style={baseStyle}
      onClick={handleClick}
      onMouseEnter={handleMouseEnter}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
        <Typography.Text strong type={disabled ? 'tertiary' : undefined}>
          {value}
        </Typography.Text>
        <Typography.Text type='secondary' size='small'>
          {label}
        </Typography.Text>
      </div>
      {item.ratio && renderRatio(item.ratio)}
    </div>
  );
};

export function renderNumber(num) {
  if (num >= 1000000000) {
    return (num / 1000000000).toFixed(1) + 'B';
  } else if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M';
  } else if (num >= 10000) {
    return (num / 1000).toFixed(1) + 'k';
  } else {
    return num;
  }
}

export function renderQuotaNumberWithDigit(num, digits = 2) {
  if (typeof num !== 'number' || isNaN(num)) {
    return 0;
  }
  let displayInCurrency = localStorage.getItem('display_in_currency');
  num = num.toFixed(digits);
  if (displayInCurrency) {
    return '$' + num;
  }
  return num;
}

export function renderNumberWithPoint(num) {
  if (num === undefined) return '';
  num = num.toFixed(2);
  if (num >= 100000) {
    // Convert number to string to manipulate it
    let numStr = num.toString();
    // Find the position of the decimal point
    let decimalPointIndex = numStr.indexOf('.');

    let wholePart = numStr;
    let decimalPart = '';

    // If there is a decimal point, split the number into whole and decimal parts
    if (decimalPointIndex !== -1) {
      wholePart = numStr.slice(0, decimalPointIndex);
      decimalPart = numStr.slice(decimalPointIndex);
    }

    // Take the first two and last two digits of the whole number part
    let shortenedWholePart = wholePart.slice(0, 2) + '..' + wholePart.slice(-2);

    // Return the formatted number
    return shortenedWholePart + decimalPart;
  }

  // If the number is less than 100,000, return it unmodified
  return num;
}

export function getQuotaPerUnit() {
  let quotaPerUnit = localStorage.getItem('quota_per_unit');
  quotaPerUnit = parseFloat(quotaPerUnit);
  return quotaPerUnit;
}

export function renderUnitWithQuota(quota) {
  let quotaPerUnit = localStorage.getItem('quota_per_unit');
  quotaPerUnit = parseFloat(quotaPerUnit);
  quota = parseFloat(quota);
  return quotaPerUnit * quota;
}

export function getQuotaWithUnit(quota, digits = 6) {
  let quotaPerUnit = localStorage.getItem('quota_per_unit');
  quotaPerUnit = parseFloat(quotaPerUnit);
  return (quota / quotaPerUnit).toFixed(digits);
}

export function renderQuotaWithAmount(amount) {
  let displayInCurrency = localStorage.getItem('display_in_currency');
  displayInCurrency = displayInCurrency === 'true';
  if (displayInCurrency) {
    return '$' + amount;
  } else {
    return renderUnitWithQuota(amount);
  }
}

export function renderQuota(quota, digits = 2) {
  let quotaPerUnit = localStorage.getItem('quota_per_unit');
  let displayInCurrency = localStorage.getItem('display_in_currency');
  quotaPerUnit = parseFloat(quotaPerUnit);
  displayInCurrency = displayInCurrency === 'true';
  if (displayInCurrency) {
    return '$' + (quota / quotaPerUnit).toFixed(digits);
  }
  return renderNumber(quota);
}

export function renderModelPrice(
  inputTokens,
  completionTokens,
  modelRatio,
  modelPrice = -1,
  completionRatio,
  groupRatio,
  cacheTokens = 0,
  cacheRatio = 1.0,
) {
  if (modelPrice !== -1) {
    return i18next.t(
      '模型价格：${{price}} * 分组倍率：{{ratio}} = ${{total}}',
      {
        price: modelPrice,
        ratio: groupRatio,
        total: modelPrice * groupRatio,
      },
    );
  } else {
    if (completionRatio === undefined) {
      completionRatio = 0;
    }
    let inputRatioPrice = modelRatio * 2.0;
    let completionRatioPrice = modelRatio * 2.0 * completionRatio;
    let cacheRatioPrice = modelRatio * 2.0 * cacheRatio;

    // Calculate effective input tokens (non-cached + cached with ratio applied)
    const effectiveInputTokens =
      inputTokens - cacheTokens + cacheTokens * cacheRatio;

    let price =
      (effectiveInputTokens / 1000000) * inputRatioPrice * groupRatio +
      (completionTokens / 1000000) * completionRatioPrice * groupRatio;

    return (
      <>
        <article>
          <p>
            {i18next.t('提示价格：${{price}} / 1M tokens', {
              price: inputRatioPrice,
            })}
          </p>
          <p>
            {i18next.t(
              '补全价格：${{price}} * {{completionRatio}} = ${{total}} / 1M tokens (补全倍率: {{completionRatio}})',
              {
                price: inputRatioPrice,
                total: completionRatioPrice,
                completionRatio: completionRatio,
              },
            )}
          </p>
          {cacheTokens > 0 && (
            <p>
              {i18next.t(
                '缓存价格：${{price}} * {{cacheRatio}} = ${{total}} / 1M tokens (缓存倍率: {{cacheRatio}})',
                {
                  price: inputRatioPrice,
                  total: inputRatioPrice * cacheRatio,
                  cacheRatio: cacheRatio,
                },
              )}
            </p>
          )}
          <p></p>
          <p>
            {cacheTokens > 0
              ? i18next.t(
                  '提示 {{nonCacheInput}} tokens / 1M tokens * ${{price}} + 缓存 {{cacheInput}} tokens / 1M tokens * ${{cachePrice}} + 补全 {{completion}} tokens / 1M tokens * ${{compPrice}} * 分组 {{ratio}} = ${{total}}',
                  {
                    nonCacheInput: inputTokens - cacheTokens,
                    cacheInput: cacheTokens,
                    cachePrice: inputRatioPrice * cacheRatio,
                    price: inputRatioPrice,
                    completion: completionTokens,
                    compPrice: completionRatioPrice,
                    ratio: groupRatio,
                    total: price.toFixed(6),
                  },
                )
              : i18next.t(
                  '提示 {{input}} tokens / 1M tokens * ${{price}} + 补全 {{completion}} tokens / 1M tokens * ${{compPrice}} * 分组 {{ratio}} = ${{total}}',
                  {
                    input: inputTokens,
                    price: inputRatioPrice,
                    completion: completionTokens,
                    compPrice: completionRatioPrice,
                    ratio: groupRatio,
                    total: price.toFixed(6),
                  },
                )}
          </p>
          <p>{i18next.t('仅供参考，以实际扣费为准')}</p>
        </article>
      </>
    );
  }
}

export function renderModelPriceSimple(
  modelRatio,
  modelPrice = -1,
  groupRatio,
  cacheTokens = 0,
  cacheRatio = 1.0,
) {
  if (modelPrice !== -1) {
    return i18next.t('价格：${{price}} * 分组：{{ratio}}', {
      price: modelPrice,
      ratio: groupRatio,
    });
  } else {
    if (cacheTokens !== 0) {
      return i18next.t(
        '模型: {{ratio}} * 分组: {{groupRatio}} * 缓存: {{cacheRatio}}',
        {
          ratio: modelRatio,
          groupRatio: groupRatio,
          cacheRatio: cacheRatio,
        },
      );
    } else {
      return i18next.t('模型: {{ratio}} * 分组: {{groupRatio}}', {
        ratio: modelRatio,
        groupRatio: groupRatio,
      });
    }
  }
}

export function renderAudioModelPrice(
  inputTokens,
  completionTokens,
  modelRatio,
  modelPrice = -1,
  completionRatio,
  audioInputTokens,
  audioCompletionTokens,
  audioRatio,
  audioCompletionRatio,
  groupRatio,
  cacheTokens = 0,
  cacheRatio = 1.0,
) {
  // 1 ratio = $0.002 / 1K tokens
  if (modelPrice !== -1) {
    return i18next.t(
      '模型价格：${{price}} * 分组倍率：{{ratio}} = ${{total}}',
      {
        price: modelPrice,
        ratio: groupRatio,
        total: modelPrice * groupRatio,
      },
    );
  } else {
    if (completionRatio === undefined) {
      completionRatio = 0;
    }

    // try toFixed audioRatio
    audioRatio = parseFloat(audioRatio).toFixed(6);
    // 这里的 *2 是因为 1倍率=0.002刀，请勿删除
    let inputRatioPrice = modelRatio * 2.0;
    let completionRatioPrice = modelRatio * 2.0 * completionRatio;
    let cacheRatioPrice = modelRatio * 2.0 * cacheRatio;

    // Calculate effective input tokens (non-cached + cached with ratio applied)
    const effectiveInputTokens =
      inputTokens - cacheTokens + cacheTokens * cacheRatio;

    let textPrice =
      (effectiveInputTokens / 1000000) * inputRatioPrice * groupRatio +
      (completionTokens / 1000000) * completionRatioPrice * groupRatio;
    let audioPrice =
      (audioInputTokens / 1000000) * inputRatioPrice * audioRatio * groupRatio +
      (audioCompletionTokens / 1000000) *
        inputRatioPrice *
        audioRatio *
        audioCompletionRatio *
        groupRatio;
    let price = textPrice + audioPrice;
    return (
      <>
        <article>
          <p>
            {i18next.t('提示价格：${{price}} / 1M tokens', {
              price: inputRatioPrice,
            })}
          </p>
          <p>
            {i18next.t(
              '补全价格：${{price}} * {{completionRatio}} = ${{total}} / 1M tokens (补全倍率: {{completionRatio}})',
              {
                price: inputRatioPrice,
                total: completionRatioPrice,
                completionRatio: completionRatio,
              },
            )}
          </p>
          {cacheTokens > 0 && (
            <p>
              {i18next.t(
                '缓存价格：${{price}} * {{cacheRatio}} = ${{total}} / 1M tokens (缓存倍率: {{cacheRatio}})',
                {
                  price: inputRatioPrice,
                  total: inputRatioPrice * cacheRatio,
                  cacheRatio: cacheRatio,
                },
              )}
            </p>
          )}
          <p>
            {i18next.t(
              '音频提示价格：${{price}} * {{audioRatio}} = ${{total}} / 1M tokens (音频倍率: {{audioRatio}})',
              {
                price: inputRatioPrice,
                total: inputRatioPrice * audioRatio,
                audioRatio: audioRatio,
              },
            )}
          </p>
          <p>
            {i18next.t(
              '音频补全价格：${{price}} * {{audioRatio}} * {{audioCompRatio}} = ${{total}} / 1M tokens (音频补全倍率: {{audioCompRatio}})',
              {
                price: inputRatioPrice,
                total: inputRatioPrice * audioRatio * audioCompletionRatio,
                audioRatio: audioRatio,
                audioCompRatio: audioCompletionRatio,
              },
            )}
          </p>
          <p>
            {cacheTokens > 0
              ? i18next.t(
                  '文字提示 {{nonCacheInput}} tokens / 1M tokens * ${{price}} + 缓存 {{cacheInput}} tokens / 1M tokens * ${{cachePrice}} + 文字补全 {{completion}} tokens / 1M tokens * ${{compPrice}} = ${{total}}',
                  {
                    nonCacheInput: inputTokens - cacheTokens,
                    cacheInput: cacheTokens,
                    cachePrice: inputRatioPrice * cacheRatio,
                    price: inputRatioPrice,
                    completion: completionTokens,
                    compPrice: completionRatioPrice,
                    total: textPrice.toFixed(6),
                  },
                )
              : i18next.t(
                  '文字提示 {{input}} tokens / 1M tokens * ${{price}} + 文字补全 {{completion}} tokens / 1M tokens * ${{compPrice}} = ${{total}}',
                  {
                    input: inputTokens,
                    price: inputRatioPrice,
                    completion: completionTokens,
                    compPrice: completionRatioPrice,
                    total: textPrice.toFixed(6),
                  },
                )}
          </p>
          <p>
            {i18next.t(
              '音频提示 {{input}} tokens / 1M tokens * ${{audioInputPrice}} + 音频补全 {{completion}} tokens / 1M tokens * ${{audioCompPrice}} = ${{total}}',
              {
                input: audioInputTokens,
                completion: audioCompletionTokens,
                audioInputPrice: audioRatio * inputRatioPrice,
                audioCompPrice:
                  audioRatio * audioCompletionRatio * inputRatioPrice,
                total: audioPrice.toFixed(6),
              },
            )}
          </p>
          <p>
            {i18next.t(
              '总价：文字价格 {{textPrice}} + 音频价格 {{audioPrice}} = ${{total}}',
              {
                total: price.toFixed(6),
                textPrice: textPrice.toFixed(6),
                audioPrice: audioPrice.toFixed(6),
              },
            )}
          </p>
          <p>{i18next.t('仅供参考，以实际扣费为准')}</p>
        </article>
      </>
    );
  }
}

export function renderQuotaWithPrompt(quota, digits) {
  let displayInCurrency = localStorage.getItem('display_in_currency');
  displayInCurrency = displayInCurrency === 'true';
  if (displayInCurrency) {
    return (
      ' | ' + i18next.t('等价金额') + ': ' + renderQuota(quota, digits) + ''
    );
  }
  return '';
}

const colors = [
  'amber',
  'blue',
  'cyan',
  'green',
  'grey',
  'indigo',
  'light-blue',
  'lime',
  'orange',
  'pink',
  'purple',
  'red',
  'teal',
  'violet',
  'yellow',
];

// 基础10色色板 (N ≤ 10)
const baseColors = [
  '#1664FF', // 主色
  '#1AC6FF',
  '#FF8A00',
  '#3CC780',
  '#7442D4',
  '#FFC400',
  '#304D77',
  '#B48DEB',
  '#009488',
  '#FF7DDA',
];

// 扩展20色色板 (10 < N ≤ 20)
const extendedColors = [
  '#1664FF',
  '#B2CFFF',
  '#1AC6FF',
  '#94EFFF',
  '#FF8A00',
  '#FFCE7A',
  '#3CC780',
  '#B9EDCD',
  '#7442D4',
  '#DDC5FA',
  '#FFC400',
  '#FAE878',
  '#304D77',
  '#8B959E',
  '#B48DEB',
  '#EFE3FF',
  '#009488',
  '#59BAA8',
  '#FF7DDA',
  '#FFCFEE',
];

export const modelColorMap = {
  'dall-e': 'rgb(147,112,219)', // 深紫色
  // 'dall-e-2': 'rgb(147,112,219)', // 介于紫色和蓝色之间的色调
  'dall-e-3': 'rgb(153,50,204)', // 介于紫罗兰和洋红之间的色调
  'gpt-3.5-turbo': 'rgb(184,227,167)', // 浅绿色
  // 'gpt-3.5-turbo-0301': 'rgb(131,220,131)', // 亮绿色
  'gpt-3.5-turbo-0613': 'rgb(60,179,113)', // 海洋绿
  'gpt-3.5-turbo-1106': 'rgb(32,178,170)', // 浅海洋绿
  'gpt-3.5-turbo-16k': 'rgb(149,252,206)', // 淡橙色
  'gpt-3.5-turbo-16k-0613': 'rgb(119,255,214)', // 淡桃
  'gpt-3.5-turbo-instruct': 'rgb(175,238,238)', // 粉蓝色
  'gpt-4': 'rgb(135,206,235)', // 天蓝色
  // 'gpt-4-0314': 'rgb(70,130,180)', // 钢蓝色
  'gpt-4-0613': 'rgb(100,149,237)', // 矢车菊蓝
  'gpt-4-1106-preview': 'rgb(30,144,255)', // 道奇蓝
  'gpt-4-0125-preview': 'rgb(2,177,236)', // 深天蓝
  'gpt-4-turbo-preview': 'rgb(2,177,255)', // 深天蓝
  'gpt-4-32k': 'rgb(104,111,238)', // 中紫色
  // 'gpt-4-32k-0314': 'rgb(90,105,205)', // 暗灰蓝色
  'gpt-4-32k-0613': 'rgb(61,71,139)', // 暗蓝灰色
  'gpt-4-all': 'rgb(65,105,225)', // 皇家蓝
  'gpt-4-gizmo-*': 'rgb(0,0,255)', // 纯蓝色
  'gpt-4-vision-preview': 'rgb(25,25,112)', // 午夜蓝
  'text-ada-001': 'rgb(255,192,203)', // 粉红色
  'text-babbage-001': 'rgb(255,160,122)', // 浅珊瑚色
  'text-curie-001': 'rgb(219,112,147)', // 苍紫罗兰色
  // 'text-davinci-002': 'rgb(199,21,133)', // 中紫罗兰红色
  'text-davinci-003': 'rgb(219,112,147)', // 苍紫罗兰色（与Curie相同，表示同一个系列）
  'text-davinci-edit-001': 'rgb(255,105,180)', // 热粉色
  'text-embedding-ada-002': 'rgb(255,182,193)', // 浅粉红
  'text-embedding-v1': 'rgb(255,174,185)', // 浅粉红色（略有区别）
  'text-moderation-latest': 'rgb(255,130,171)', // 强粉色
  'text-moderation-stable': 'rgb(255,160,122)', // 浅珊瑚色（与Babbage相同，表示同一类功能）
  'tts-1': 'rgb(255,140,0)', // 深橙色
  'tts-1-1106': 'rgb(255,165,0)', // 橙色
  'tts-1-hd': 'rgb(255,215,0)', // 金色
  'tts-1-hd-1106': 'rgb(255,223,0)', // 金黄色（略有区别）
  'whisper-1': 'rgb(245,245,220)', // 米色
  'claude-3-opus-20240229': 'rgb(255,132,31)', // 橙红色
  'claude-3-sonnet-20240229': 'rgb(253,135,93)', // 橙色
  'claude-3-haiku-20240307': 'rgb(255,175,146)', // 浅橙色
  'claude-2.1': 'rgb(255,209,190)', // 浅橙色（略有区别）
};

export function modelToColor(modelName) {
  // 1. 如果模型在预定义的 modelColorMap 中，使用预定义颜色
  if (modelColorMap[modelName]) {
    return modelColorMap[modelName];
  }

  // 2. 生成一个稳定的数字作为索引
  let hash = 0;
  for (let i = 0; i < modelName.length; i++) {
    hash = (hash << 5) - hash + modelName.charCodeAt(i);
    hash = hash & hash; // Convert to 32-bit integer
  }
  hash = Math.abs(hash);

  // 3. 根据模型名称长度选择不同的色板
  const colorPalette = modelName.length > 10 ? extendedColors : baseColors;

  // 4. 使用hash值选择颜色
  const index = hash % colorPalette.length;
  return colorPalette[index];
}

export function stringToColor(str) {
  let sum = 0;
  for (let i = 0; i < str.length; i++) {
    sum += str.charCodeAt(i);
  }
  let i = sum % colors.length;
  return colors[i];
}

export function renderClaudeModelPrice(
  inputTokens,
  completionTokens,
  modelRatio,
  modelPrice = -1,
  completionRatio,
  groupRatio,
  cacheTokens = 0,
  cacheRatio = 1.0,
  cacheCreationTokens = 0,
  cacheCreationRatio = 1.0,
) {
  const ratioLabel = false ? i18next.t('专属倍率') : i18next.t('分组倍率');

  if (modelPrice !== -1) {
    return i18next.t(
      '模型价格：${{price}} * {{ratioType}}：{{ratio}} = ${{total}}',
      {
        price: modelPrice,
        ratioType: ratioLabel,
        ratio: groupRatio,
        total: modelPrice * groupRatio,
      },
    );
  } else {
    if (completionRatio === undefined) {
      completionRatio = 0;
    }

    const completionRatioValue = completionRatio || 0;
    const inputRatioPrice = modelRatio * 2.0;
    const completionRatioPrice = modelRatio * 2.0 * completionRatioValue;
    let cacheRatioPrice = (modelRatio * 2.0 * cacheRatio).toFixed(2);
    let cacheCreationRatioPrice = modelRatio * 2.0 * cacheCreationRatio;

    // Calculate effective input tokens (non-cached + cached with ratio applied + cache creation with ratio applied)
    const nonCachedTokens = inputTokens;
    const effectiveInputTokens =
      nonCachedTokens +
      cacheTokens * cacheRatio +
      cacheCreationTokens * cacheCreationRatio;

    let price =
      (effectiveInputTokens / 1000000) * inputRatioPrice * groupRatio +
      (completionTokens / 1000000) * completionRatioPrice * groupRatio;

    return (
      <>
        <article>
          <p>
            {i18next.t('提示价格：${{price}} / 1M tokens', {
              price: inputRatioPrice,
            })}
          </p>
          <p>
            {i18next.t(
              '补全价格：${{price}} * {{ratio}} = ${{total}} / 1M tokens',
              {
                price: inputRatioPrice,
                ratio: completionRatio,
                total: completionRatioPrice,
              },
            )}
          </p>
          {cacheTokens > 0 && (
            <p>
              {i18next.t(
                '缓存价格：${{price}} * {{ratio}} = ${{total}} / 1M tokens (缓存倍率: {{cacheRatio}})',
                {
                  price: inputRatioPrice,
                  ratio: cacheRatio,
                  total: cacheRatioPrice,
                  cacheRatio: cacheRatio,
                },
              )}
            </p>
          )}
          {cacheCreationTokens > 0 && (
            <p>
              {i18next.t(
                '缓存创建价格：${{price}} * {{ratio}} = ${{total}} / 1M tokens (缓存创建倍率: {{cacheCreationRatio}})',
                {
                  price: inputRatioPrice,
                  ratio: cacheCreationRatio,
                  total: cacheCreationRatioPrice,
                  cacheCreationRatio: cacheCreationRatio,
                },
              )}
            </p>
          )}
          <p></p>
          <p>
            {cacheTokens > 0 || cacheCreationTokens > 0
              ? i18next.t(
                  '提示 {{nonCacheInput}} tokens / 1M tokens * ${{price}} + 缓存 {{cacheInput}} tokens / 1M tokens * ${{cachePrice}} + 缓存创建 {{cacheCreationInput}} tokens / 1M tokens * ${{cacheCreationPrice}} + 补全 {{completion}} tokens / 1M tokens * ${{compPrice}} * 分组 {{ratio}} = ${{total}}',
                  {
                    nonCacheInput: nonCachedTokens,
                    cacheInput: cacheTokens,
                    cacheRatio: cacheRatio,
                    cacheCreationInput: cacheCreationTokens,
                    cacheCreationRatio: cacheCreationRatio,
                    cachePrice: cacheRatioPrice,
                    cacheCreationPrice: cacheCreationRatioPrice,
                    price: inputRatioPrice,
                    completion: completionTokens,
                    compPrice: completionRatioPrice,
                    ratio: groupRatio,
                    total: price.toFixed(6),
                  },
                )
              : i18next.t(
                  '提示 {{input}} tokens / 1M tokens * ${{price}} + 补全 {{completion}} tokens / 1M tokens * ${{compPrice}} * 分组 {{ratio}} = ${{total}}',
                  {
                    input: inputTokens,
                    price: inputRatioPrice,
                    completion: completionTokens,
                    compPrice: completionRatioPrice,
                    ratio: groupRatio,
                    total: price.toFixed(6),
                  },
                )}
          </p>
          <p>{i18next.t('仅供参考，以实际扣费为准')}</p>
        </article>
      </>
    );
  }
}

export function renderClaudeLogContent(
  modelRatio,
  completionRatio,
  modelPrice = -1,
  groupRatio,
  cacheRatio = 1.0,
  cacheCreationRatio = 1.0,
) {
  const ratioLabel = false ? i18next.t('专属倍率') : i18next.t('分组倍率');

  if (modelPrice !== -1) {
    return i18next.t('模型价格 ${{price}}，{{ratioType}} {{ratio}}', {
      price: modelPrice,
      ratioType: ratioLabel,
      ratio: groupRatio,
    });
  } else {
    return i18next.t(
      '模型倍率 {{modelRatio}}，补全倍率 {{completionRatio}}，缓存倍率 {{cacheRatio}}，缓存创建倍率 {{cacheCreationRatio}}，{{ratioType}} {{ratio}}',
      {
        modelRatio: modelRatio,
        completionRatio: completionRatio,
        cacheRatio: cacheRatio,
        cacheCreationRatio: cacheCreationRatio,
        ratioType: ratioLabel,
        ratio: groupRatio,
      },
    );
  }
}

export function renderClaudeModelPriceSimple(
  modelRatio,
  modelPrice = -1,
  groupRatio,
  cacheTokens = 0,
  cacheRatio = 1.0,
  cacheCreationTokens = 0,
  cacheCreationRatio = 1.0,
) {
  const ratioLabel = false ? i18next.t('专属倍率') : i18next.t('分组');

  if (modelPrice !== -1) {
    return i18next.t('价格：${{price}} * {{ratioType}}：{{ratio}}', {
      price: modelPrice,
      ratioType: ratioLabel,
      ratio: groupRatio,
    });
  } else {
    if (cacheTokens !== 0 || cacheCreationTokens !== 0) {
      return i18next.t(
        '模型: {{ratio}} * {{ratioType}}: {{groupRatio}} * 缓存: {{cacheRatio}}',
        {
          ratio: modelRatio,
          ratioType: ratioLabel,
          groupRatio: groupRatio,
          cacheRatio: cacheRatio,
          cacheCreationRatio: cacheCreationRatio,
        },
      );
    } else {
      return i18next.t('模型: {{ratio}} * {{ratioType}}: {{groupRatio}}', {
        ratio: modelRatio,
        ratioType: ratioLabel,
        groupRatio: groupRatio,
      });
    }
  }
}

export function renderLogContent(
  modelRatio,
  completionRatio,
  modelPrice = -1,
  groupRatio,
) {
  const ratioLabel = false ? i18next.t('专属倍率') : i18next.t('分组倍率');

  if (modelPrice !== -1) {
    return i18next.t('模型价格 ${{price}}，{{ratioType}} {{ratio}}', {
      price: modelPrice,
      ratioType: ratioLabel,
      ratio: groupRatio,
    });
  } else {
    return i18next.t(
      '模型倍率 {{modelRatio}}，补全倍率 {{completionRatio}}，{{ratioType}} {{ratio}}',
      {
        modelRatio: modelRatio,
        completionRatio: completionRatio,
        ratioType: ratioLabel,
        ratio: groupRatio,
      },
    );
  }
}
