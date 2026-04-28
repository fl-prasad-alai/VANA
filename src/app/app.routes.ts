import { Routes } from '@angular/router';
import { LoginComponent } from './pages/login/login.component';
import { RegisterComponent } from './pages/register/register.component';
import { ForestGuideComponent } from './components/forest-guide/forest-guide.component';
import { authGuard } from './guards/auth.guard';

export const routes: Routes = [
  {
    path: '',
    redirectTo: 'dashboard',
    pathMatch: 'full'
  },
  {
    path: 'login',
    component: LoginComponent
  },
  {
    path: 'register',
    component: RegisterComponent
  },
  {
    path: 'dashboard',
    component: ForestGuideComponent,
    canActivate: [authGuard]
  },
  {
    path: '**',
    redirectTo: 'dashboard'
  }
];
