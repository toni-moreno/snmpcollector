import { Routes } from '@angular/router';
import { Home } from '../home/home';
import { Login } from '../login/login';
import { SwaggerUiComponent } from '../swagger-ui/swagger-ui.component';

export const AppRoutes: Routes = [
  { path: '', redirectTo: '/home', pathMatch: 'full' },
  { path: 'home', component: Home },
  { path: 'sign-in', component: Login },
  { path: 'docs', component: SwaggerUiComponent },
  ];
