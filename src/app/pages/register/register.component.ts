// src/app/pages/register/register.component.ts

import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { CustomValidators, getPasswordStrengthError } from '../../utils/validators';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  template: `
<div
  class="flex items-center justify-center min-h-screen transition-colors duration-500 py-8"
  [ngClass]="isDarkMode ? 'dark bg-moss-dark' : 'bg-moss-light'"
>
  <!-- Background Decoration -->
  <div
    class="absolute top-0 right-0 w-96 h-96 rounded-full opacity-10 blur-3xl"
    [ngClass]="isDarkMode ? 'bg-moss-accent' : 'bg-moss-primary'"
  ></div>
  <div
    class="absolute bottom-0 left-0 w-96 h-96 rounded-full opacity-10 blur-3xl"
    [ngClass]="isDarkMode ? 'bg-moss-primary' : 'bg-moss-accent'"
  ></div>

  <!-- Register Card -->
  <div class="relative w-full max-w-md px-6 py-8">
    <!-- Header -->
    <div class="text-center mb-8">
      <h1
        class="text-4xl font-bold tracking-wider mb-2"
        [ngClass]="
          isDarkMode
            ? 'text-moss-accent'
            : 'text-moss-primary'
        "
      >
        🌿 Emerald Moss
      </h1>
      <p
        class="text-sm"
        [ngClass]="
          isDarkMode ? 'text-moss-muted' : 'text-moss-secondary'
        "
      >
        Begin Your Journey
      </p>
    </div>

    <!-- Theme Toggle -->
    <div class="flex justify-center mb-8">
      <button
        (click)="toggleTheme()"
        class="px-4 py-2 rounded-full transition-all glass-panel hover"
        [ngClass]="isDarkMode ? 'dark' : ''"
      >
        {{ isDarkMode ? '☀️ Light Mode' : '🌙 Dark Mode' }}
      </button>
    </div>

    <!-- Alert Messages -->
    <div *ngIf="errorMessage" class="alert alert-error mb-4">
      <span>⚠️</span>
      <div>{{ errorMessage }}</div>
    </div>

    <div *ngIf="successMessage" class="alert alert-success mb-4">
      <span>✓</span>
      <div>{{ successMessage }}</div>
    </div>

    <!-- Register Form Card -->
    <div class="glass-panel" [ngClass]="isDarkMode ? 'dark' : ''">
      <form [formGroup]="registerForm" (ngSubmit)="onSubmit()" class="space-y-5">
        <!-- Full Name Field -->
        <div class="form-group">
          <label for="fullName" class="form-label">Full Name</label>
          <input
            id="fullName"
            type="text"
            class="form-input"
            [ngClass]="{
              'error': registerForm.get('fullName')?.invalid && registerForm.get('fullName')?.touched,
              'dark': isDarkMode
            }"
            formControlName="fullName"
            placeholder="John Doe"
            [disabled]="isLoading"
          />
          <div
            *ngIf="
              registerForm.get('fullName')?.invalid &&
              registerForm.get('fullName')?.touched
            "
            class="form-error"
          >
            Full name is required
          </div>
        </div>

        <!-- Email Field -->
        <div class="form-group">
          <label for="email" class="form-label">Email Address</label>
          <input
            id="email"
            type="email"
            class="form-input"
            [ngClass]="{
              'error': registerForm.get('email')?.invalid && registerForm.get('email')?.touched,
              'dark': isDarkMode
            }"
            formControlName="email"
            placeholder="your@email.com"
            [disabled]="isLoading"
          />
          <div
            *ngIf="
              registerForm.get('email')?.invalid &&
              registerForm.get('email')?.touched
            "
            class="form-error"
          >
            {{ getEmailError() }}
          </div>
        </div>

        <!-- Password Field -->
        <div class="form-group">
          <label for="password" class="form-label">Password</label>
          <input
            id="password"
            type="password"
            class="form-input"
            [ngClass]="{
              'error': registerForm.get('password')?.invalid && registerForm.get('password')?.touched,
              'dark': isDarkMode
            }"
            formControlName="password"
            placeholder="••••••••"
            [disabled]="isLoading"
          />
          <div
            *ngIf="
              registerForm.get('password')?.invalid &&
              registerForm.get('password')?.touched
            "
            class="form-error"
          >
            {{ getPasswordError() }}
          </div>
          <div class="form-hint">
            Password must contain uppercase, lowercase, number, special character, and be at least 8 characters
          </div>
        </div>

        <!-- Confirm Password Field -->
        <div class="form-group">
          <label for="confirmPassword" class="form-label">Confirm Password</label>
          <input
            id="confirmPassword"
            type="password"
            class="form-input"
            [ngClass]="{
              'error': registerForm.get('confirmPassword')?.invalid && registerForm.get('confirmPassword')?.touched,
              'dark': isDarkMode
            }"
            formControlName="confirmPassword"
            placeholder="••••••••"
            [disabled]="isLoading"
          />
          <div
            *ngIf="
              registerForm.get('confirmPassword')?.invalid &&
              registerForm.get('confirmPassword')?.touched
            "
            class="form-error"
          >
            {{ getConfirmPasswordError() }}
          </div>
        </div>

        <!-- Terms Agreement -->
        <div class="flex items-start gap-3">
          <input
            id="agreeTerms"
            type="checkbox"
            class="w-4 h-4 rounded cursor-pointer mt-1"
            [ngClass]="isDarkMode ? 'dark' : ''"
            formControlName="agreeTerms"
            [disabled]="isLoading"
          />
          <label
            for="agreeTerms"
            class="text-sm cursor-pointer"
            [ngClass]="
              isDarkMode ? 'text-moss-muted' : 'text-moss-secondary'
            "
          >
            I agree to the
            <a href="#" class="underline hover:opacity-80">Terms of Service</a>
            and
            <a href="#" class="underline hover:opacity-80">Privacy Policy</a>
          </label>
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          class="btn btn-primary w-full"
          [ngClass]="isDarkMode ? 'dark' : ''"
          [disabled]="isLoading || registerForm.invalid"
        >
          <span *ngIf="!isLoading">Create Account</span>
          <span *ngIf="isLoading" class="flex items-center gap-2">
            <div class="spinner small"></div>
            Creating...
          </span>
        </button>
      </form>

      <!-- Divider -->
      <div
        class="my-6 flex items-center gap-4"
        [ngClass]="isDarkMode ? 'text-moss-muted' : 'text-moss-secondary'"
      >
        <div class="flex-1 h-px bg-current opacity-30"></div>
        <span class="text-xs">Already registered?</span>
        <div class="flex-1 h-px bg-current opacity-30"></div>
      </div>

      <!-- Login Link -->
      <div class="text-center">
        <a
          routerLink="/login"
          class="btn btn-secondary w-full"
          [ngClass]="isDarkMode ? 'dark' : ''"
        >
          Sign In Instead
        </a>
      </div>
    </div>

    <!-- Footer -->
    <div
      class="text-center mt-8 text-xs"
      [ngClass]="isDarkMode ? 'text-moss-muted' : 'text-moss-secondary'"
    >
      <p>🌿 Your mental health journey starts with a single step.</p>
    </div>
  </div>
</div>
  `,
  styles: []
})
export class RegisterComponent implements OnInit {
  registerForm!: FormGroup;
  isDarkMode = true;
  isLoading = false;
  errorMessage = '';
  successMessage = '';

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.initializeForm();
    this.setTheme();
  }

  private initializeForm(): void {
    this.registerForm = this.fb.group(
      {
        fullName: ['', [Validators.required, Validators.minLength(2)]],
        email: ['', [Validators.required, CustomValidators.email()]],
        password: ['', [Validators.required, CustomValidators.passwordStrength()]],
        confirmPassword: ['', Validators.required],
        agreeTerms: [false, Validators.requiredTrue]
      },
      { validators: CustomValidators.passwordsMatch('password', 'confirmPassword') }
    );
  }

  private setTheme(): void {
    if (this.isDarkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }

  toggleTheme(): void {
    this.isDarkMode = !this.isDarkMode;
    this.setTheme();
  }

  onSubmit(): void {
    if (this.registerForm.invalid) {
      this.markFormGroupTouched(this.registerForm);
      return;
    }

    this.isLoading = true;
    this.errorMessage = '';
    this.successMessage = '';

    const { email, password, fullName } = this.registerForm.value;

    this.authService.register(email, password, fullName).subscribe({
      next: (response) => {
        this.isLoading = false;
        if (response.success) {
          this.successMessage = 'Account created successfully! Redirecting...';
          setTimeout(() => {
            this.router.navigate(['/dashboard']);
          }, 1500);
        }
      },
      error: (error) => {
        this.isLoading = false;
        this.errorMessage = error.error?.message || 'Registration failed. Please try again.';
      }
    });
  }

  private markFormGroupTouched(formGroup: FormGroup): void {
    Object.keys(formGroup.controls).forEach(key => {
      const control = formGroup.get(key);
      control?.markAsTouched();
      if (control instanceof FormGroup) {
        this.markFormGroupTouched(control);
      }
    });
  }

  getEmailError(): string {
    const control = this.registerForm.get('email');
    if (control?.hasError('required')) {
      return 'Email is required';
    }
    if (control?.hasError('invalidEmail')) {
      return 'Please enter a valid email address';
    }
    return '';
  }

  getPasswordError(): string {
    const control = this.registerForm.get('password');
    return getPasswordStrengthError(control?.errors || null);
  }

  getConfirmPasswordError(): string {
    const control = this.registerForm.get('confirmPassword');
    if (control?.hasError('required')) {
      return 'Please confirm your password';
    }
    if (control?.hasError('passwordMismatch')) {
      return 'Passwords do not match';
    }
    return '';
  }
}
