import { Routes } from '@angular/router';
import { Home } from '../home/home';
import { Login } from '../login/login';


export const routes: Routes = [
  { path: '',   component: Login },
  { path: 'login',  component: Login },
  { path: 'home',   component: Home },
];
