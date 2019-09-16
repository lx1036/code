import {subtract} from './substract';

describe('my test', () => {
  it('should be true', () => {
    expect(true).toBe(true);
  });

  it('subtracts 2 numbers', () => {
    expect(subtract(2, 4)).toBe(-2);
  });
});
