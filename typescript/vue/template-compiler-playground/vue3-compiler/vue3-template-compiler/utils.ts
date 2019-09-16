



export function makeMap(str: string, lowerCase?: boolean): (key: string) => true| void {
  const map = Object.create(null);
  const list: string[] = str.split(',');
  
  for (let i = 0; i < list.length; i++) {
    map[list[i]] = true;
  }
  
  return  lowerCase ? val => map[val.toLowerCase()] : val => map[val];
}

