
import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { HttpAPI} from '../common/httpAPI'

@Component({
  selector: 'login',
  templateUrl: './login.html',
  styleUrls: ['./login.css']
})

export class Login {
  constructor(public router: Router, public httpAPI: HttpAPI) {
  }
  ifErrors: any;

  login(event, username, password) {
    event.preventDefault();
    let body = JSON.stringify({ username, password });
    this.httpAPI.post('/login', body)
      .subscribe(
        response => {
          this.router.navigate(['home']);
        },
        error => {
          this.ifErrors = error['_body'];
          console.log(error.text());
        }
      );
  }
}
