import { Component } from 'angular2/core';
import { Router, RouterLink } from 'angular2/router';
import { CORE_DIRECTIVES, FORM_DIRECTIVES } from 'angular2/common';
import { Http, Headers } from 'angular2/http';
import { contentHeaders } from '../common/headers';

@Component({
  selector: 'login',
  directives: [RouterLink, CORE_DIRECTIVES, FORM_DIRECTIVES ],
  templateUrl: '/public/login/login.html',
  styleUrls: ['/public/login/login.css']
})
export class Login {
  constructor(public router: Router, public http: Http) {
  }

  login(event, username, password) {
    event.preventDefault();
    let body = JSON.stringify({ username, password });
    this.http.post('/session/create', body, { headers: contentHeaders })
      .subscribe(
        response => {
          localStorage.setItem('jwt', response.json().id_token);
          this.router.parent.navigateByUrl('/');
        },
        error => {
          alert(error.text());
          console.log(error.text());
        }
      );
  }
/*
  signup(event) {
    event.preventDefault();
    this.router.parent.navigateByUrl('/signup');
  }*/
}
