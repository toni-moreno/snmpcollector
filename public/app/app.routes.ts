import { Routes } from '@angular/router';
import { Home } from '../home/home';
import { Login } from '../login/login';
import { AuthGuard } from '../common/auth.guard';

export const routes: Routes = [
  { path: '',   component: Login },
  { path: 'login',  component: Login },
  { path: 'home',   component: Home, canActivate: [AuthGuard] },
];
