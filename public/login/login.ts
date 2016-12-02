
import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { HttpAPI} from '../common/httpAPI'

@Component({
  selector: 'login',
  templateUrl: 'public/login/login.html',
  styleUrls: ['public/login/login.css']
})

export class Login {
  constructor(public router: Router, public httpAPI: HttpAPI) {
  }

  login(event, username, password) {
    event.preventDefault();
    let body = JSON.stringify({ username, password });
    this.httpAPI.post('/login', body)
      .subscribe(
        response => {
          this.router.navigate(['home']);
        },
        error => {
          alert(error.text());
          console.log(error.text());
        }
      );
  }
}
