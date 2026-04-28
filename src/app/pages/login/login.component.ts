// src/app/pages/login/login.component.ts

import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { CustomValidators } from '../../utils/validators';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  template: `
<div
  class="flex items-center justify-center min-h-screen transition-colors duration-500"
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

  <!-- Login Card -->
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
        Mental Health Companion
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

    <!-- Login Form Card -->
    <div class="glass-panel" [ngClass]="isDarkMode ? 'dark' : ''">
      <form [formGroup]="loginForm" (ngSubmit)="onSubmit()" class="space-y-6">
        <!-- Email Field -->
        <div class="form-group">
          <label for="email" class="form-label"
            >Email Address</label
          >
          <input
            id="email"
            type="email"
            class="form-input"
            [ngClass]="{
              'error': loginForm.get('email')?.invalid && loginForm.get('email')?.touched,
              'dark': isDarkMode
            }"
            formControlName="email"
            placeholder="your@email.com"
            [disabled]="isLoading"
          />
          <div
            *ngIf="
              loginForm.get('email')?.invalid &&
              loginForm.get('email')?.touched
            "
            class="form-error"
          >
            {{ getEmailError() }}
          </div>
        </div>

        <!-- Password Field -->
        <div class="form-group">
          <label for="password" class="form-label"
            >Password</label
          >
          <input
            id="password"
            type="password"
            class="form-input"
            [ngClass]="{
              'error': loginForm.get('password')?.invalid && loginForm.get('password')?.touched,
              'dark': isDarkMode
            }"
            formControlName="password"
            placeholder="••••••••"
            [disabled]="isLoading"
          />
          <div
            *ngIf="
              loginForm.get('password')?.invalid &&
              loginForm.get('password')?.touched
            "
            class="form-error"
          >
            {{ getPasswordError() }}
          </div>
        </div>

        <!-- Remember Me -->
        <div class="flex items-center gap-2">
          <input
            id="rememberMe"
            type="checkbox"
            class="w-4 h-4 rounded cursor-pointer"
            [ngClass]="isDarkMode ? 'dark' : ''"
            formControlName="rememberMe"
            [disabled]="isLoading"
          />
          <label
            for="rememberMe"
            class="text-sm cursor-pointer"
            [ngClass]="
              isDarkMode ? 'text-moss-muted' : 'text-moss-secondary'
            "
          >
            Remember me
          </label>
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          class="btn btn-primary w-full"
          [ngClass]="isDarkMode ? 'dark' : ''"
          [disabled]="isLoading || loginForm.invalid"
        >
          <span *ngIf="!isLoading">Sign In</span>
          <span *ngIf="isLoading" class="flex items-center gap-2">
            <div class="spinner small"></div>
            Signing in...
          </span>
        </button>
      </form>

      <!-- Divider -->
      <div
        class="my-6 flex items-center gap-4"
        [ngClass]="isDarkMode ? 'text-moss-muted' : 'text-moss-secondary'"
      >
        <div class="flex-1 h-px bg-current opacity-30"></div>
        <span class="text-xs">New to Emerald Moss?</span>
        <div class="flex-1 h-px bg-current opacity-30"></div>
      </div>

      <!-- Register Link -->
      <div class="text-center">
        <p
          class="text-sm mb-4"
          [ngClass]="
            isDarkMode ? 'text-moss-muted' : 'text-moss-secondary'
          "
        >
          Don't have an account?
        </p>
        <a
          routerLink="/register"
          class="btn btn-secondary w-full"
          [ngClass]="isDarkMode ? 'dark' : ''"
        >
          Create Account
        </a>
      </div>
    </div>

    <!-- Footer -->
    <div
      class="text-center mt-8 text-xs"
      [ngClass]="isDarkMode ? 'text-moss-muted' : 'text-moss-secondary'"
    >
      <p class="mb-4">
        By signing in, you agree to our
        <a href="#" class="underline hover:opacity-80">Terms of Service</a>
        and
        <a href="#" class="underline hover:opacity-80">Privacy Policy</a>
      </p>
      <p>🌿 Your mental health journey starts here.</p>
    </div>
  </div>
</div>
  `,
  styles: []
})
export class LoginComponent implements OnInit {
  loginForm!: FormGroup;
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
    this.loginForm = this.fb.group({
      email: ['', [Validators.required, CustomValidators.email()]],
      password: ['', [Validators.required, Validators.minLength(8)]],
      rememberMe: [false]
    });
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
    if (this.loginForm.invalid) {
      this.markFormGroupTouched(this.loginForm);
      return;
    }

    this.isLoading = true;
    this.errorMessage = '';
    this.successMessage = '';

    const { email, password } = this.loginForm.value;

    this.authService.login(email, password).subscribe({
      next: (response) => {
        this.isLoading = false;
        if (response.success) {
          this.successMessage = 'Login successful! Redirecting...';
          setTimeout(() => {
            this.router.navigate(['/dashboard']);
          }, 1000);
        }
      },
      error: (error) => {
        this.isLoading = false;
        this.errorMessage = error.error?.message || 'Login failed. Please try again.';
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
    const control = this.loginForm.get('email');
    if (control?.hasError('required')) {
      return 'Email is required';
    }
    if (control?.hasError('invalidEmail')) {
      return 'Please enter a valid email address';
    }
    return '';
  }

  getPasswordError(): string {
    const control = this.loginForm.get('password');
    if (control?.hasError('required')) {
      return 'Password is required';
    }
    if (control?.hasError('minlength')) {
      return 'Password must be at least 8 characters';
    }
    return '';
  }
}

