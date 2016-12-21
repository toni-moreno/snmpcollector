import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'snmpcol-app',
  templateUrl: './app.html'
})

export class App {
  constructor(public router: Router) {}
}
