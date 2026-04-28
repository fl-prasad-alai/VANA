import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { BreathSyncComponent } from '../breath-sync/breath-sync.component';

@Component({
  selector: 'app-forest-guide',
  standalone: true,
  imports: [CommonModule, BreathSyncComponent],
  template: `
    <div class="flex flex-col h-screen max-w-2xl mx-auto p-4 transition-colors duration-500 ease-in-out" 
         [ngClass]="isDarkMode ? 'dark bg-moss-dark' : 'bg-moss-light'">
         
      <!-- Header -->
      <header class="flex justify-between items-center py-4 mb-8 glass-panel px-6">
        <h1 class="text-2xl font-bold text-moss-primary dark:text-moss-accent tracking-wider">Emerald Moss</h1>
        <button (click)="toggleTheme()" class="p-2 rounded-full hover:bg-moss-glass transition">
          <span *ngIf="isDarkMode">🌿 Light Mode</span>
          <span *ngIf="!isDarkMode">🌲 Dark Mode</span>
        </button>
      </header>

      <!-- Chat Area -->
      <div class="flex-1 overflow-y-auto space-y-4 mb-4">
        <!-- AI Message -->
        <div class="flex flex-col space-y-2 max-w-[85%]">
          <div class="glass-panel p-4 rounded-tl-none animate-fade-in text-moss-dark dark:text-moss-light">
            Welcome to the Emerald Moss. I am here to guide you. Take a deep breath. How are you feeling today?
          </div>
        </div>
      </div>

      <!-- Voice Interaction Area -->
      <div class="mt-auto flex justify-center py-8">
        <app-breath-sync [isListening]="isListening" (click)="toggleListening()"></app-breath-sync>
      </div>
      
      <div class="text-center text-sm text-moss-secondary dark:text-moss-accent opacity-70">
        {{ isListening ? 'Listening... speak naturally.' : 'Tap to speak.' }}
      </div>
    </div>
  `,
  styles: []
})
export class ForestGuideComponent {
  isDarkMode = true;
  isListening = true;

  toggleTheme() {
    this.isDarkMode = !this.isDarkMode;
    if (this.isDarkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }

  toggleListening() {
    this.isListening = !this.isListening;
  }
}
