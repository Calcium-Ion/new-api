export function getLogOther(otherStr) {
  if (otherStr === undefined || otherStr === '') {
    otherStr = '{}';
  }
  let other = JSON.parse(otherStr);
  return other;
}
