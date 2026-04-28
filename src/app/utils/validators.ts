// src/app/utils/validators.ts

import { AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';

export class CustomValidators {
  // Email format validator
  static email(): ValidatorFn {
    return (control: AbstractControl): ValidationErrors | null => {
      if (!control.value) {
        return null;
      }
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return emailRegex.test(control.value) ? null : { invalidEmail: true };
    };
  }

  // Password strength validator
  static passwordStrength(): ValidatorFn {
    return (control: AbstractControl): ValidationErrors | null => {
      if (!control.value) {
        return null;
      }
      const password = control.value;
      const hasUpperCase = /[A-Z]/.test(password);
      const hasLowerCase = /[a-z]/.test(password);
      const hasNumeric = /[0-9]/.test(password);
      const hasSpecialChar = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password);
      const isLongEnough = password.length >= 8;

      const passwordValid = hasUpperCase && hasLowerCase && hasNumeric && hasSpecialChar && isLongEnough;

      if (!passwordValid) {
        return {
          weakPassword: {
            hasUpperCase,
            hasLowerCase,
            hasNumeric,
            hasSpecialChar,
            isLongEnough
          }
        };
      }
      return null;
    };
  }

  // Passwords match validator
  static passwordsMatch(passwordField: string, confirmPasswordField: string): ValidatorFn {
    return (formGroup: AbstractControl): ValidationErrors | null => {
      const passwordControl = formGroup.get(passwordField);
      const confirmPasswordControl = formGroup.get(confirmPasswordField);

      if (!passwordControl || !confirmPasswordControl) {
        return null;
      }

      if (confirmPasswordControl.errors && !confirmPasswordControl.errors['passwordMismatch']) {
        return null;
      }

      if (passwordControl.value !== confirmPasswordControl.value) {
        confirmPasswordControl.setErrors({ passwordMismatch: true });
        return { passwordMismatch: true };
      } else {
        confirmPasswordControl.setErrors(null);
        return null;
      }
    };
  }
}

export const getPasswordStrengthError = (errors: ValidationErrors | null): string => {
  if (!errors) return '';
  if (errors['required']) return 'Password is required';
  if (errors['weakPassword']) {
    const err = errors['weakPassword'];
    const missing = [];
    if (!err.hasUpperCase) missing.push('uppercase letter');
    if (!err.hasLowerCase) missing.push('lowercase letter');
    if (!err.hasNumeric) missing.push('number');
    if (!err.hasSpecialChar) missing.push('special character');
    if (!err.isLongEnough) missing.push('at least 8 characters');
    return `Password must contain: ${missing.join(', ')}`;
  }
  return '';
};
