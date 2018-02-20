import { shallow } from 'enzyme';
import {
  noNeg,
  capitalizeFirstLetter,
  formatBytes, formatPrice,
  transformProductsBarChart, transformProductsPieChart,
  getTotalPieChart
} from '../formatters';

describe('Formatters', () => {

  describe('NoNeg', () => {

    const validInput = 42;
    const negativeInput = -42;
    const stringInput = "aa";
    const stringNegativeInput = "-42";

    it('should return input value', () => {
      expect(noNeg(validInput)).toBe(validInput);
      expect(noNeg(stringInput)).toBe(stringInput);
    });

    it('should return zero', () => {
      expect(noNeg(negativeInput)).toBe(0);
      expect(noNeg(stringNegativeInput)).toBe(0);
    });

  });

  describe('CapitalizeFirstLetter', () => {

    const validInput = "Test";
    const invalidInput = "test";

    it('should return input value', () => {
      expect(capitalizeFirstLetter(validInput)).toBe(validInput);
    });

    it('should return capitalized value', () => {
      expect(capitalizeFirstLetter(invalidInput)).toBe(validInput);
    });

  });

  describe('FormatBytes', () => {

    const validInput = 42;
    const zeroValue = "0 Bytes";

    it('should return formatted value', () => {
      expect(formatBytes(validInput)).toBe(validInput + " Bytes");
    });

    it('should return zero value', () => {
      expect(formatBytes(0)).toBe(zeroValue);
    });

  });

  describe('FormatPrice', () => {

    const validInput = 42.042;
    const formattedValue = "42.04";

    it('should return formatted value', () => {
      const output = shallow(formatPrice(validInput));
      expect(output.length).toBe(1);
      const spans = output.find("span");
      expect(spans.length).toBe(2);
      expect(spans.first().props().children[1]).toBe(formattedValue);
    });

  });

  describe('TansformProductsBarChart', () => {

    const days = {
      day: {
        day1: 42,
        day2: 21
      }
    };

    const costsByProductPerDay = {
      product: {
        product1: {...days},
        product2: {...days}
      }
    };

    const costsAll = {...days};

    const costsMissingDays = {
      product: {
        product1: {...days},
        product2: {
          day: {
            ...days.day,
            day3: 84
          }
        },
      }
    };

    const costsMissingKeys = {
      product: {
        ...costsByProductPerDay.product,
        "": days
      }
    };

    it('returns an empty array when invalid filter', () => {
      expect(transformProductsBarChart(costsByProductPerDay, "region", "day")).toEqual([]);
    });

    it('returns an empty array when valid filter and invalid interval', () => {
      expect(transformProductsBarChart(costsByProductPerDay, "product", "month")).toEqual([]);
    });

    it('returns an empty array when filter is "all" and invalid interval', () => {
      expect(transformProductsBarChart(costsAll, "all", "month")).toEqual([]);
    });

    it('returns formatted array when valid filter and valid interval', () => {
      const output = [{
        key: "product1",
        values: [["day1", days.day.day1], ["day2", days.day.day2]]
      },{
        key: "product2",
        values: [["day1", days.day.day1], ["day2", days.day.day2]]
      }];
      expect(transformProductsBarChart(costsByProductPerDay, "product", "day")).toEqual(output);
    });

    it('returns formatted array when filter is "all" and valid interval', () => {
      const output = [{
        key: "Total",
        values: [["day1", days.day.day1], ["day2", days.day.day2]]
      }];
      expect(transformProductsBarChart(costsAll, "all", "day")).toEqual(output);
    });

    it('fills missing days', () => {
      const output = [{
        key: "product1",
        values: [["day1", days.day.day1], ["day2", days.day.day2], ["day3", 0]]
      },{
        key: "product2",
        values: [["day1", days.day.day1], ["day2", days.day.day2], ["day3", costsMissingDays.product.product2.day.day3]]
      }];
      expect(transformProductsBarChart(costsMissingDays, "product", "day")).toEqual(output);
    });

    it('fills missing keys', () => {
      const output = [{
        key: "product1",
        values: [["day1", days.day.day1], ["day2", days.day.day2]]
      },{
        key: "product2",
        values: [["day1", days.day.day1], ["day2", days.day.day2]]
      },{
        key: "No product",
        values: [["day1", days.day.day1], ["day2", days.day.day2]]
      }];
      expect(transformProductsBarChart(costsMissingKeys, "product", "day")).toEqual(output);
    });

  });

  describe('TransformProductsPieChart', () => {

    const costsByProduct = {
      product: {
        product1: 42,
        product2: 84
      }
    };

    it('returns an empty array when invalid filter', () => {
      expect(transformProductsPieChart(costsByProduct, "region")).toEqual([]);
    });

    it('returns formatted array when valid filter and valid interval', () => {
      const output = [{
        key: "product1",
        value: costsByProduct.product.product1
      },{
        key: "product2",
        value: costsByProduct.product.product2
      }];
      expect(transformProductsPieChart(costsByProduct, "product")).toEqual(output);
    });

  });

  describe('GetTotalPieChart', () => {

    const data = [{
      key: "product1",
      value: 42
    },{
      key: "product2",
      value: 84
    }];

    it('returns an empty array when invalid data', () => {
      expect(getTotalPieChart(42)).toEqual(0);
    });

    it('returns total when valid data', () => {
      expect(getTotalPieChart(data)).toEqual((42 +  84));
    });

  });

});
