import {Pipe, PipeTransform} from '@angular/core';

@Pipe({ name: 'elapsedseconds' })
export class ElapsedSecondsPipe implements PipeTransform {
  transform(value: any, ...args: string[]): string {
    if (typeof args === 'undefined' || args.length !== 1) {
      throw new Error('ElapsedSecondsPipe: missing required decimals');
    }

    return this.toSeconds(value, args[0] ,0);
  }
  //from kbn.js in grafana project
  toFixed(value , decimals )  {
    if (value === null) {
      return "";
    }

    var factor = decimals ? Math.pow(10, Math.max(0, decimals)) : 1;
    var formatted = String(Math.round(value * factor) / factor);

    // if exponent return directly
    if (formatted.indexOf('e') !== -1 || value === 0) {
      return formatted;
    }

    // If tickDecimals was specified, ensure that we have exactly that
    // much precision; otherwise default to the value's own precision.
    if (decimals != null) {
      var decimalPos = formatted.indexOf(".");
      var precision = decimalPos === -1 ? 0 : formatted.length - decimalPos - 1;
      if (precision < decimals) {
        return (precision ? formatted : formatted + ".") + (String(factor)).substr(1, decimals - precision);
      }
    }

    return formatted;
  }

  toFixedScaled(value, decimals , scaledDecimals , additionalDecimals , ext ) {
    if (scaledDecimals === null) {
        return this.toFixed(value, decimals) + ext;
    } else {
       return this.toFixed(value, scaledDecimals + additionalDecimals) + ext;
     }
  }

  toSeconds(size , decimals, scaledDecimals) {
    if (size === null) { return ""; }

    // Less than 1 µs, devide in ns
    if (Math.abs(size) < 0.000001) {
      return this.toFixedScaled(size * 1.e9, decimals, scaledDecimals - decimals, -9, " ns");
    }
    // Less than 1 ms, devide in µs
    if (Math.abs(size) < 0.001) {
      return this.toFixedScaled(size * 1.e6, decimals, scaledDecimals - decimals, -6, " µs");
    }
    // Less than 1 second, devide in ms
    if (Math.abs(size) < 1) {
      return this.toFixedScaled(size * 1.e3, decimals, scaledDecimals - decimals, -3, " ms");
    }

    if (Math.abs(size) < 60) {
      return this.toFixed(size, decimals) + " s";
    }
    // Less than 1 hour, devide in minutes
    else if (Math.abs(size) < 3600) {
      return this.toFixedScaled(size / 60, decimals, scaledDecimals, 1, " min");
    }
    // Less than one day, devide in hours
    else if (Math.abs(size) < 86400) {
      return this.toFixedScaled(size / 3600, decimals, scaledDecimals, 4, " hour");
    }
    // Less than one week, devide in days
    else if (Math.abs(size) < 604800) {
      return this.toFixedScaled(size / 86400, decimals, scaledDecimals, 5, " day");
    }
    // Less than one year, devide in week
    else if (Math.abs(size) < 31536000) {
      return this.toFixedScaled(size / 604800, decimals, scaledDecimals, 6, " week");
    }

    return this.toFixedScaled(size / 3.15569e7, decimals, scaledDecimals, 7, " year");
  }
}
