import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-breath-sync',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="flex items-center justify-center p-8">
      <div 
        class="relative flex items-center justify-center w-32 h-32 rounded-full shadow-2xl glass-panel"
        [ngClass]="{'animate-breath-sync': isListening}"
      >
        <div class="absolute w-full h-full rounded-full border-4 border-moss-primary opacity-30"></div>
        <div class="absolute w-3/4 h-3/4 rounded-full bg-moss-accent opacity-50 blur-sm"></div>
        <div class="z-10 text-moss-primary dark:text-moss-light">
          <svg *ngIf="isListening" xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
          </svg>
          <svg *ngIf="!isListening" xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4" />
          </svg>
        </div>
      </div>
    </div>
  `,
  styles: []
})
export class BreathSyncComponent {
  @Input() isListening: boolean = true;
}
