import {Component} from 'angular2/core';
import {Location, RouteConfig, RouterLink, Router} from 'angular2/router';
//import {Location} from 'angular2/platform/common';

import {LoggedInRouterOutlet} from './LoggedInOutlet';
import {Home} from '../home/home';
import {Login} from '../login/login';


@Component({
  selector: 'auth-app',
  templateUrl: 'public/app/app.html',
  directives: [ LoggedInRouterOutlet ]
})
@RouteConfig([
  { path: '/', redirectTo: ['/Home'] },
  { path: '/home', component: Home, as: 'Home' },
  { path: '/login', component: Login, as: 'Login' }
//  { path: '/signup', component: Signup, as: 'Signup' }
])

export class App {
  constructor(public router: Router) {
  }
}
