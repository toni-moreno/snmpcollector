import {Component, Input, OnDestroy} from '@angular/core';

@Component({
  selector: 'my-spinner',
  template: `
    <div [hidden]="!isDelayedRunning" class="sk-circle">
      <div>
        <div class="sk-circle1 sk-child"></div>
          <div class="sk-circle2 sk-child"></div>
          <div class="sk-circle3 sk-child"></div>
          <div class="sk-circle4 sk-child"></div>
          <div class="sk-circle5 sk-child"></div>
          <div class="sk-circle6 sk-child"></div>
          <div class="sk-circle7 sk-child"></div>
          <div class="sk-circle8 sk-child"></div>
          <div class="sk-circle9 sk-child"></div>
          <div class="sk-circle10 sk-child"></div>
          <div class="sk-circle11 sk-child"></div>
          <div class="sk-circle12 sk-child"></div>
        </div>
      <div style="display: block; margin-left: 60px; padding-top: 5px; width: 200px">
        <span *ngIf="message">{{message}}</span>
      </div>
    </div>  `,
  styleUrls: ['../css/spinner.css']


})
export class SpinnerComponent implements OnDestroy {
  private currentTimeout: any;
  public isDelayedRunning: boolean = false;

  @Input()
  public delay: number = 100;

  @Input() message: string;
  @Input()
  public set isRunning(value: boolean) {
    if (!value) {
      this.cancelTimeout();
      this.isDelayedRunning = false;
      return;
    }

    if (this.currentTimeout) {
      return;
    }

    this.currentTimeout = setTimeout(() => {
      this.isDelayedRunning = value;
      this.cancelTimeout();
    }, this.delay);
  }

  private cancelTimeout(): void {
    clearTimeout(this.currentTimeout);
    this.currentTimeout = undefined;
  }

  ngOnDestroy(): any {
    this.cancelTimeout();
  }
}
