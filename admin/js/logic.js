define(function() {
	var $password = $('#password');
	var $newurl = $('#newurl');
	var $addbtn = $('#add');
	var $activate = $('#activate');
	var $table = $('#table');
	var $body = $($table.find('tbody'));
	var $notificationarea = $('#notificationarea');
	var $templates = $('.template');
	var $successtpl = $($templates.filter('.alert-success'));
	var $errortpl = $($templates.filter('.alert-error'));
	var $rowtpl = $($templates.filter('tr'));

	var timeToLive = function(obj, time) {
		setTimeout(function() {
			$(obj).remove();
		}, time*1000);
	}

	var obj = {
		setState: function(state, password) {
			var oldstate = state != 'active'?'checked':false;
			$.ajax({
				'url': '/api/state',
				'data': state,
				'type': 'POST',
				'headers': {
					'X-DiplomaEnhancer-Token': password,
				},
				'statusCode': {
					400: function() {
						this.showError('Invalid state');
						$activate.attr('disabled');

					}.bind(this),
					401: function() {
						this.showError('Wrong password');
						$activate.attr('disabled');
					}.bind(this),
					204: function() {
						this.showSuccess('DiplomaEnhancer is now '+state);
						$activate.attr('disabled');
					}.bind(this),
				},
				error: function() {
					$activate.attr('checked', oldstate);
				}
			});
		},
		showSuccess: function(msg) {
			var $success = $($successtpl.clone().removeClass('template'));
			$($success.find('.text')).text(msg);
			$success.appendTo($notificationarea);
			timeToLive($success, 5);
		},
		showError: function(msg) {
			var $error = $($errortpl.clone().removeClass('template'));
			$($error.find('.text')).text(msg);
			$error.appendTo($notificationarea);
			timeToLive($error, 5);
		},
		clearTable: function() {
			$table.find('tbody tr').each(function() {
				if(!$(this).hasClass('template')) {
					$(this).remove();
				}
			});
		},
		addRow: function(url) {
			var $row = $($rowtpl.clone().removeClass('template'));
			$row.find('button').click(function(){
				var password = $password.attr('value');
				this.unblock(url, password);
			}.bind(this));
			$row.children('.url').text(url);
			$row.appendTo($body);
		},
		removeRow: function(url) {
			$($body.find('tr')).each(function() {
				if($($(this).find('.url')).text() == url) {
					$(this).remove();
				}
			});
		},
		updateTable: function() {
			this.clearTable();
			$.ajax({
				'url': '/api/',
				'dataType': 'json',
				'success': function(data) {
					for(var ip in data) {
						for(var urlidx in data[ip]) {
							var url = data[ip][urlidx]
							if(url.match(/host$/)) {
								continue;
							}
							this.addRow(url);
						}
					}
				}.bind(this),
				'error': function() {
					this.showError('Could not load table');
				}.bind(this),
			});
		},
		block: function(url, password) {
			$.ajax({
				'url': '/api/127.0.0.1',
				'data': url,
				'type': 'POST',
				'headers': {
					'X-DiplomaEnhancer-Token': password,
				},
				'statusCode': {
					401: function() {
						this.showError('Wrong password');
					}.bind(this),
					204: function() {
						this.addRow(url);
						this.showSuccess('Blocked '+url);
						$newurl.attr('value', '');
					}.bind(this),
				},
			});
		},
		unblock: function(url, password) {
			$.ajax({
				'url': '/api/127.0.0.1',
				'data': url,
				'type': 'DELETE',
				'headers': {
					'X-DiplomaEnhancer-Token': password,
				},
				'statusCode': {
					400: function() {
						this.showError('Invalid operation');
					}.bind(this),
					401: function() {
						this.showError('Wrong password');
					}.bind(this),
					204: function() {
						this.removeRow(url);
						this.showSuccess('Unblocked '+url);
					}.bind(this),
				},
			});
		},
	};

	$addbtn.click(function() {
		var url = $newurl.attr('value');
		var password = $password.attr('value');
		obj.block(url, password);
	});

	$activate.click(function() {
		console.log('drin');
		$activate.attr('disabled');
		var state = $activate.attr('checked')?'active':'inactive';
		var password = $password.attr('value');
		obj.setState(state, password);
	})
	return obj;
});
