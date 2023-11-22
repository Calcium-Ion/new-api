import { Label } from 'semantic-ui-react';

export function renderText(text, limit) {
  if (text.length > limit) {
    return text.slice(0, limit - 3) + '...';
  }
  return text;
}

export function renderGroup(group) {
  if (group === '') {
    return <Label>default</Label>;
  }
  let groups = group.split(',');
  groups.sort();
  return <>
    {groups.map((group) => {
      if (group === 'vip' || group === 'pro') {
        return <Label color='yellow'>{group}</Label>;
      } else if (group === 'svip' || group === 'premium') {
        return <Label color='red'>{group}</Label>;
      }
      return <Label>{group}</Label>;
    })}
  </>;
}

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

export function getQuotaPerUnit() {
  let quotaPerUnit = localStorage.getItem('quota_per_unit');
  quotaPerUnit = parseFloat(quotaPerUnit);
  return quotaPerUnit;
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

export function renderQuotaWithPrompt(quota, digits) {
  let displayInCurrency = localStorage.getItem('display_in_currency');
  displayInCurrency = displayInCurrency === 'true';
  if (displayInCurrency) {
    return `（等价金额：${renderQuota(quota, digits)}）`;
  }
  return '';
}

const colors = ['amber', 'blue', 'cyan', 'green', 'grey', 'indigo',
  'light-blue', 'lime', 'orange', 'pink',
  'purple', 'red', 'teal', 'violet', 'yellow'
]

export function stringToColor(str) {
  let sum = 0;
  // 对字符串中的每个字符进行操作
  for (let i = 0; i < str.length; i++) {
    // 将字符的ASCII值加到sum中
    sum += str.charCodeAt(i);
  }
  // 使用模运算得到个位数
  let i = sum % colors.length;
  return colors[i];
}