import {sum} from "./index";

describe('sum in jasmine', () => {
  it('sum', () => {
    expect(sum(1,2)).toEqual(3);
  });
  
  it('sum2', () => {
    expect(sum(2,2)).toEqual(4);
  });
});
