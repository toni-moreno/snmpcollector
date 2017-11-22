
import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { HttpService } from '../core/http.service'

@Component({
  selector: 'login',
  templateUrl: './login.html',
  styleUrls: ['./login.css']
})

export class Login {
  constructor(public router: Router, public httpAPI: HttpService) {
    this.getFooterInfo();
  }
    public ifErrors: any;
    public version : any;
    public userIn : boolean = false;

  login(event, username, password) {
    event.preventDefault();
    let body = JSON.stringify({ username, password });
    this.httpAPI.post('/login', body, null, true)
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

    getInfo() {
        // return an observable
        return this.httpAPI.get('/api/rt/agent/info/version/')
        .map( (responseData) => responseData.json())
    }

    getFooterInfo() {
      this.getInfo()
      .subscribe(data => {
        this.version = data;
        this.userIn = true;
      },
      err => console.error(err),
      () =>  {}
      );
    }
}
