import { render } from 'enzyme';
import Validation from '../AuthForm';

describe('Authentication Form Validation', () => {

  describe('Required Validation', () => {

    it('should render no error', () => {
      expect(Validation.required("test")).toBe(undefined);
    });

    it('should render an error', () => {
      const result = render(Validation.required(""));
      expect(result.hasClass("alert alert-warning")).toBe(true);
    });

  });

  describe('Email Validation', () => {

    it('should render no error', () => {
      expect(Validation.email("test@test.test")).toBe(undefined);
    });

    it('should render an error', () => {
      const result = render(Validation.email("test"));
      expect(result.hasClass("alert alert-warning")).toBe(true);
    });

  });

});