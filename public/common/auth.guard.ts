import { Injectable } from '@angular/core';
import { Router, CanActivate } from '@angular/router';
import { tokenNotExpired } from 'angular2-jwt';


@Injectable()
export class AuthGuard implements CanActivate {
  constructor(private router: Router) {}

  canActivate() {
    console.log('check canActivate:');
    var token: string;
    token=localStorage.getItem('id_token');
    if (token) {
          return true;
    }

    /*
    * THIS method crashes until not valid JWT token has been released returnig true
    if (tokenNotExpired()) {
      return true;
    }*/

    this.router.navigate(['login']);
    return false;
  }
}
